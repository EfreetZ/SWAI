package patterns

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPoolSubmitAndClose(t *testing.T) {
	pool, err := NewWorkerPool(4, 16)
	if err != nil {
		t.Fatalf("NewWorkerPool() error = %v", err)
	}

	var count atomic.Int32
	var wg sync.WaitGroup
	taskCount := 20
	wg.Add(taskCount)

	for i := 0; i < taskCount; i++ {
		err = pool.Submit(context.Background(), func(ctx context.Context) error {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				count.Add(1)
				return nil
			}
		})
		if err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("tasks did not finish in time")
	}

	if got := count.Load(); got != int32(taskCount) {
		t.Fatalf("count = %d, want %d", got, taskCount)
	}

	closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := pool.Close(closeCtx); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if err := pool.Submit(context.Background(), func(ctx context.Context) error { return nil }); err != ErrPoolClosed {
		t.Fatalf("Submit() after Close = %v, want %v", err, ErrPoolClosed)
	}
}

func TestWorkerPoolRunTimeout(t *testing.T) {
	pool, err := NewWorkerPool(1, 1)
	if err != nil {
		t.Fatalf("NewWorkerPool() error = %v", err)
	}
	defer func() {
		_ = pool.Close(context.Background())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err = pool.Run(ctx, func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	})
	if err == nil {
		t.Fatal("Run() error = nil, want timeout error")
	}
}

func TestWorkerPoolInvalidArgs(t *testing.T) {
	if _, err := NewWorkerPool(0, 1); err != ErrInvalidWorkerCount {
		t.Fatalf("worker count error = %v, want %v", err, ErrInvalidWorkerCount)
	}
	if _, err := NewWorkerPool(1, 0); err != ErrInvalidQueueSize {
		t.Fatalf("queue size error = %v, want %v", err, ErrInvalidQueueSize)
	}
}

func BenchmarkWorkerPoolRun(b *testing.B) {
	pool, err := NewWorkerPool(8, 128)
	if err != nil {
		b.Fatalf("NewWorkerPool() error = %v", err)
	}
	defer func() {
		_ = pool.Close(context.Background())
	}()

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := pool.Run(ctx, func(ctx context.Context) error { return nil }); err != nil {
			b.Fatalf("Run() error = %v", err)
		}
	}
}
