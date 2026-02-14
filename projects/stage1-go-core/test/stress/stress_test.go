package stress

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage1-go-core/datastructures/linkedlist"
)

// TestLRUStressConcurrentPutGet 通过高并发读写验证 LRU 的稳定性。
func TestLRUStressConcurrentPutGet(t *testing.T) {
	cache, err := linkedlist.NewLRUCache[string, int](2048)
	if err != nil {
		t.Fatalf("NewLRUCache() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const workers = 64
	const loops = 2000

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < loops; i++ {
				select {
				case <-ctx.Done():
					return
				default:
					key := "w" + strconv.Itoa(workerID) + "-" + strconv.Itoa(i)
					cache.Put(key, i)
					_, _ = cache.Get(key)
				}
			}
		}(w)
	}

	wg.Wait()
	if cache.Len() > 2048 {
		t.Fatalf("Len() = %d, want <= 2048", cache.Len())
	}
}
