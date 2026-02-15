package kv

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/raft"
)

func TestPutGet(t *testing.T) {
	cluster := raft.NewCluster([]string{"n1", "n2", "n3"})
	svc := NewService(cluster)
	if err := svc.Put(context.Background(), "k", "v"); err != nil {
		t.Fatalf("put failed: %v", err)
	}
	v, ok := svc.Get("k")
	if !ok || v != "v" {
		t.Fatalf("unexpected get: %v %v", v, ok)
	}
}
