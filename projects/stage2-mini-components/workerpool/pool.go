package workerpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrInvalidWorkers = errors.New("max workers must be greater than 0")
	ErrInvalidQueue   = errors.New("queue size must be greater than 0")
	ErrPoolClosed     = errors.New("worker pool is closed")
	ErrTaskNil        = errors.New("task must not be nil")
)

// Task 表示可执行任务。
type Task func(context.Context) error

type request struct {
	ctx    context.Context
	task   Task
	result chan error
}

// Pool 是带优雅关闭能力的工作池。
type Pool struct {
	queue   chan request
	closeCh chan struct{}
	closed  atomic.Bool
	once    sync.Once
	wg      sync.WaitGroup

	running atomic.Int64
}

// New 创建并启动工作池。
func New(maxWorkers, queueSize int) (*Pool, error) {
	if maxWorkers <= 0 {
		return nil, ErrInvalidWorkers
	}
	if queueSize <= 0 {
		return nil, ErrInvalidQueue
	}

	p := &Pool{queue: make(chan request, queueSize), closeCh: make(chan struct{})}
	for i := 0; i < maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
	return p, nil
}

// Submit 提交任务并等待执行结果。
func (p *Pool) Submit(ctx context.Context, task Task) error {
	if task == nil {
		return ErrTaskNil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if p.closed.Load() {
		return ErrPoolClosed
	}

	result := make(chan error, 1)
	req := request{ctx: ctx, task: task, result: result}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.closeCh:
		return ErrPoolClosed
	case p.queue <- req:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-result:
		return err
	}
}

// SubmitWithTimeout 带超时提交任务。
func (p *Pool) SubmitWithTimeout(task Task, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return p.Submit(ctx, task)
}

// Running 返回活跃 worker 数。
func (p *Pool) Running() int {
	return int(p.running.Load())
}

// Shutdown 优雅关闭，等待 worker 退出。
func (p *Pool) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	p.once.Do(func() {
		p.closed.Store(true)
		close(p.closeCh)
	})

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (p *Pool) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.closeCh:
			return
		case req := <-p.queue:
			p.running.Add(1)
			err := req.task(req.ctx)
			p.running.Add(-1)
			if req.result != nil {
				req.result <- err
			}
		}
	}
}
