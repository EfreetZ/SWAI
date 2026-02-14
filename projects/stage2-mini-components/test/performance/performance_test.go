package performance

import (
	"context"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage2-mini-components/workerpool"
)

// TestWorkerPoolPerformance 验证任务池在固定负载下的吞吐表现。
func TestWorkerPoolPerformance(t *testing.T) {
	pool, err := workerpool.New(16, 1024)
	if err != nil {
		t.Fatalf("workerpool.New() error = %v", err)
	}
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	const total = 5000
	start := time.Now()
	for i := 0; i < total; i++ {
		err = pool.Submit(context.Background(), func(ctx context.Context) error {
			return nil
		})
		if err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
	}

	elapsed := time.Since(start)
	throughput := float64(total) / elapsed.Seconds()
	t.Logf("workerpool throughput=%.2f ops/s elapsed=%s", throughput, elapsed)
}
