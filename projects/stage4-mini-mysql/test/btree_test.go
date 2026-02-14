package test

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
)

func TestBTreeRangeWithManyKeys(t *testing.T) {
	tree := storage.NewBPlusTree(32, nil)
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		key := []byte{byte(i)}
		if err := tree.Insert(ctx, key, key); err != nil {
			t.Fatalf("Insert() error = %v", err)
		}
	}

	it, err := tree.RangeScan(ctx, []byte{10}, []byte{20})
	if err != nil {
		t.Fatalf("RangeScan() error = %v", err)
	}
	count := 0
	for it.Next() {
		count++
	}
	if count != 11 {
		t.Fatalf("count = %d, want 11", count)
	}
}
