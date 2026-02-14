package patterns

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	// ErrInvalidWorkerCount 表示 worker 数量非法。
	ErrInvalidWorkerCount = errors.New("worker count must be greater than 0")
	// ErrInvalidQueueSize 表示队列长度非法。
	ErrInvalidQueueSize = errors.New("queue size must be greater than 0")
	// ErrInvalidTask 表示任务为空。
	ErrInvalidTask = errors.New("task must not be nil")
	// ErrPoolClosed 表示任务池已关闭。
	ErrPoolClosed = errors.New("worker pool is closed")
)

// Task 表示可被调度执行的任务函数。
type Task func(context.Context) error

type taskRequest struct {
	ctx    context.Context
	task   Task
	result chan error
}

// WorkerPool 提供受控并发执行能力。
type WorkerPool struct {
	tasks     chan taskRequest
	closeCh   chan struct{}
	closeOnce sync.Once
	closed    atomic.Bool
	wg        sync.WaitGroup
}

// NewWorkerPool 创建并启动固定数量的 worker。
func NewWorkerPool(workerCount, queueSize int) (*WorkerPool, error) {
	if workerCount <= 0 {
		return nil, ErrInvalidWorkerCount
	}
	if queueSize <= 0 {
		return nil, ErrInvalidQueueSize
	}

	pool := &WorkerPool{
		tasks:   make(chan taskRequest, queueSize),
		closeCh: make(chan struct{}),
	}

	for i := 0; i < workerCount; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool, nil
}

// Submit 异步提交任务，不等待执行结果。
func (p *WorkerPool) Submit(ctx context.Context, task Task) error {
	if task == nil {
		return ErrInvalidTask
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if p.closed.Load() {
		return ErrPoolClosed
	}

	request := taskRequest{ctx: ctx, task: task}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.closeCh:
		return ErrPoolClosed
	case p.tasks <- request:
		return nil
	}
}

// Run 同步提交任务，并等待任务执行完成。
func (p *WorkerPool) Run(ctx context.Context, task Task) error {
	if task == nil {
		return ErrInvalidTask
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if p.closed.Load() {
		return ErrPoolClosed
	}

	resultCh := make(chan error, 1)
	request := taskRequest{ctx: ctx, task: task, result: resultCh}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.closeCh:
		return ErrPoolClosed
	case p.tasks <- request:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-resultCh:
		return err
	}
}

// Close 停止任务池并等待所有 worker 退出。
func (p *WorkerPool) Close(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	p.closeOnce.Do(func() {
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

func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.closeCh:
			return
		case req := <-p.tasks:
			err := req.task(req.ctx)
			if req.result != nil {
				req.result <- err
			}
		}
	}
}
