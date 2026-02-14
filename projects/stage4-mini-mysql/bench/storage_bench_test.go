package bench

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
)

func BenchmarkStorageInsertSearch(b *testing.B) {
	pager, err := storage.NewFilePageManager(filepath.Join(b.TempDir(), "bench.db"))
	if err != nil {
		b.Fatalf("NewFilePageManager() error = %v", err)
	}
	defer func() { _ = pager.Close() }()

	tree := storage.NewBPlusTree(16, pager)
	ctx := context.Background()

	b.Run("insert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := []byte{byte(i), byte(i >> 8), byte(i >> 16)}
			if err := tree.Insert(ctx, key, key); err != nil {
				b.Fatalf("Insert() error = %v", err)
			}
		}
	})

	for i := 0; i < 10000; i++ {
		key := []byte{byte(i), byte(i >> 8), byte(i >> 16)}
		_ = tree.Insert(ctx, key, key)
	}

	b.Run("search", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := []byte{byte(i % 10000), byte((i % 10000) >> 8), byte((i % 10000) >> 16)}
			if _, err := tree.Search(ctx, key); err != nil {
				b.Fatalf("Search() error = %v", err)
			}
		}
	})
}
