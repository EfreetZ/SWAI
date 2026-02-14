package performance

import (
	"context"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage1-go-core/concurrency/patterns"
)

// TestWorkerPoolPerformance 用于验证任务池在固定负载下的吞吐表现。
func TestWorkerPoolPerformance(t *testing.T) {
	pool, err := patterns.NewWorkerPool(16, 512)
	if err != nil {
		t.Fatalf("NewWorkerPool() error = %v", err)
	}
	defer func() {
		_ = pool.Close(context.Background())
	}()

	const totalTasks = 2000
	start := time.Now()

	for i := 0; i < totalTasks; i++ {
		err = pool.Run(context.Background(), func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	}

	elapsed := time.Since(start)
	throughput := float64(totalTasks) / elapsed.Seconds()
	t.Logf("worker_pool throughput=%.2f ops/s elapsed=%s", throughput, elapsed)
}
