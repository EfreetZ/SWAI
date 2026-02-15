package performance

import (
	"context"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/kv"
	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/raft"
)

func TestKVPerformance(t *testing.T) {
	cluster := raft.NewCluster([]string{"n1", "n2", "n3"})
	svc := kv.NewService(cluster)
	ctx := context.Background()
	start := time.Now()
	for i := 0; i < 5000; i++ {
		if err := svc.Put(ctx, "k", "v"); err != nil {
			t.Fatalf("put failed: %v", err)
		}
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("performance regression: %s", time.Since(start))
	}
}
