package chaos

import (
	"context"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage1-go-core/concurrency/patterns"
)

// TestWorkerPoolChaosCancel 模拟任务执行中途取消，验证系统可恢复。
func TestWorkerPoolChaosCancel(t *testing.T) {
	pool, err := patterns.NewWorkerPool(4, 32)
	if err != nil {
		t.Fatalf("NewWorkerPool() error = %v", err)
	}
	defer func() {
		_ = pool.Close(context.Background())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err = pool.Run(ctx, func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return nil
		}
	})
	if err == nil {
		t.Fatal("Run() error = nil, want context canceled/deadline exceeded")
	}

	// 验证取消场景后任务池仍然可继续处理新任务。
	err = pool.Run(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Run() after chaos error = %v", err)
	}
}
