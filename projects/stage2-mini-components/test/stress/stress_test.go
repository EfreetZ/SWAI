package stress

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage2-mini-components/lru"
	"github.com/EfreetZ/SWAI/projects/stage2-mini-components/workerpool"
)

// TestStressConcurrentCacheAndPool 压力验证缓存与任务池在高并发下稳定运行。
func TestStressConcurrentCacheAndPool(t *testing.T) {
	cache, err := lru.NewConcurrent[string, int](4096)
	if err != nil {
		t.Fatalf("lru.NewConcurrent() error = %v", err)
	}
	pool, err := workerpool.New(32, 2048)
	if err != nil {
		t.Fatalf("workerpool.New() error = %v", err)
	}
	defer func() {
		_ = pool.Shutdown(context.Background())
	}()

	const workers = 128
	const loops = 500
	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < loops; i++ {
				key := "w" + strconv.Itoa(workerID) + "-" + strconv.Itoa(i)
				cache.Put(key, i)
				_ = pool.Submit(context.Background(), func(ctx context.Context) error {
					_, _ = cache.Get(key)
					return nil
				})
			}
		}(w)
	}
	wg.Wait()

	if cache.Len() == 0 {
		t.Fatal("cache should not be empty after stress test")
	}
}
