package stress

import (
	"context"
	"sync"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/kv"
	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/raft"
)

func TestKVStress(t *testing.T) {
	cluster := raft.NewCluster([]string{"n1", "n2", "n3"})
	svc := kv.NewService(cluster)
	ctx := context.Background()

	const workers = 20
	const each = 300
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < each; i++ {
				_ = svc.Put(ctx, "k", "v")
			}
		}()
	}
	wg.Wait()
}
