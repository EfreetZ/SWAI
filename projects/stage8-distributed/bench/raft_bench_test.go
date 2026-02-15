package bench

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/kv"
	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/raft"
)

func BenchmarkPut(b *testing.B) {
	cluster := raft.NewCluster([]string{"n1", "n2", "n3"})
	svc := kv.NewService(cluster)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.Put(ctx, "k", "v")
	}
}
