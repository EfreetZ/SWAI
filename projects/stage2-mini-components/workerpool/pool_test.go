package workerpool

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolSubmitAndShutdown(t *testing.T) {
	pool, err := New(4, 16)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var count atomic.Int32
	for i := 0; i < 20; i++ {
		err = pool.Submit(context.Background(), func(ctx context.Context) error {
			count.Add(1)
			return nil
		})
		if err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
	}
	if count.Load() != 20 {
		t.Fatalf("count = %d, want 20", count.Load())
	}

	if err = pool.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	err = pool.Submit(context.Background(), func(ctx context.Context) error { return nil })
	if err != ErrPoolClosed {
		t.Fatalf("Submit() after shutdown = %v, want %v", err, ErrPoolClosed)
	}
}

func TestPoolSubmitTimeout(t *testing.T) {
	pool, err := New(1, 1)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	err = pool.SubmitWithTimeout(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}, 10*time.Millisecond)
	if err == nil {
		t.Fatal("SubmitWithTimeout() error = nil, want timeout")
	}
}

func BenchmarkPoolSubmit(b *testing.B) {
	pool, err := New(8, 256)
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	for i := 0; i < b.N; i++ {
		if submitErr := pool.Submit(context.Background(), func(ctx context.Context) error { return nil }); submitErr != nil {
			b.Fatalf("Submit() error = %v", submitErr)
		}
	}
}
