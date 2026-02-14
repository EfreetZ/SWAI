package delayqueue

import (
	"container/heap"
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrTaskNil      = errors.New("task must not be nil")
	ErrTaskIDEmpty  = errors.New("task id must not be empty")
	ErrTaskExists   = errors.New("task id already exists")
	ErrTaskNotFound = errors.New("task not found")
)

// Task 表示一个延迟执行任务。
type Task struct {
	ID        string
	ExecuteAt time.Time
	Callback  func()
	index     int
}

type taskHeap []*Task

func (h taskHeap) Len() int           { return len(h) }
func (h taskHeap) Less(i, j int) bool { return h[i].ExecuteAt.Before(h[j].ExecuteAt) }
func (h taskHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *taskHeap) Push(x any)        { n := x.(*Task); n.index = len(*h); *h = append(*h, n) }
func (h *taskHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	item.index = -1
	*h = old[:n-1]
	return item
}
func (h taskHeap) Peek() *Task {
	if len(h) == 0 {
		return nil
	}
	return h[0]
}

// HeapQueue 是最小堆延迟队列。
type HeapQueue struct {
	mu     sync.Mutex
	tasks  taskHeap
	index  map[string]*Task
	notify chan struct{}
}

// NewHeapQueue 创建延迟队列。
func NewHeapQueue() *HeapQueue {
	q := &HeapQueue{
		index:  make(map[string]*Task),
		notify: make(chan struct{}, 1),
	}
	heap.Init(&q.tasks)
	return q
}

// Add 添加延迟任务。
func (q *HeapQueue) Add(task *Task) error {
	if task == nil || task.Callback == nil {
		return ErrTaskNil
	}
	if task.ID == "" {
		return ErrTaskIDEmpty
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	if _, exists := q.index[task.ID]; exists {
		return ErrTaskExists
	}
	heap.Push(&q.tasks, task)
	q.index[task.ID] = task
	q.signal()
	return nil
}

// Cancel 取消一个任务。
func (q *HeapQueue) Cancel(taskID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	t, ok := q.index[taskID]
	if !ok {
		return ErrTaskNotFound
	}
	heap.Remove(&q.tasks, t.index)
	delete(q.index, taskID)
	q.signal()
	return nil
}

// Start 启动调度循环，直到 ctx 取消。
func (q *HeapQueue) Start(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	for {
		nextWait := q.nextWaitDuration()
		if nextWait < 0 {
			select {
			case <-ctx.Done():
				return
			case <-q.notify:
			}
			continue
		}

		timer := time.NewTimer(nextWait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return
		case <-q.notify:
			if !timer.Stop() {
				<-timer.C
			}
		case <-timer.C:
			q.runDueTasks()
		}
	}
}

func (q *HeapQueue) runDueTasks() {
	for {
		q.mu.Lock()
		peek := q.tasks.Peek()
		if peek == nil || peek.ExecuteAt.After(time.Now()) {
			q.mu.Unlock()
			return
		}
		t := heap.Pop(&q.tasks).(*Task)
		delete(q.index, t.ID)
		q.mu.Unlock()

		go t.Callback()
	}
}

func (q *HeapQueue) nextWaitDuration() time.Duration {
	q.mu.Lock()
	defer q.mu.Unlock()
	peek := q.tasks.Peek()
	if peek == nil {
		return -1
	}
	wait := time.Until(peek.ExecuteAt)
	if wait < 0 {
		return 0
	}
	return wait
}

func (q *HeapQueue) signal() {
	select {
	case q.notify <- struct{}{}:
	default:
	}
}
