package chaos

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage2-mini-components/circuitbreaker"
	"github.com/EfreetZ/SWAI/projects/stage2-mini-components/workerpool"
)

// TestChaosBreakerAndPool 模拟下游故障与任务超时，验证系统可恢复。
func TestChaosBreakerAndPool(t *testing.T) {
	cb := circuitbreaker.New(2, 1, 20*time.Millisecond, 1)
	pool, err := workerpool.New(2, 8)
	if err != nil {
		t.Fatalf("workerpool.New() error = %v", err)
	}
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	fnErr := errors.New("downstream failed")
	_ = cb.Execute(func() error { return fnErr })
	_ = cb.Execute(func() error { return fnErr })

	if state := cb.State(); state != circuitbreaker.Open {
		t.Fatalf("state = %v, want Open", state)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err = pool.Submit(ctx, func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	})
	if err == nil {
		t.Fatal("Submit() error = nil, want timeout")
	}

	time.Sleep(25 * time.Millisecond)
	if err = cb.Execute(func() error { return nil }); err != nil {
		t.Fatalf("breaker recover execute error = %v", err)
	}
}
