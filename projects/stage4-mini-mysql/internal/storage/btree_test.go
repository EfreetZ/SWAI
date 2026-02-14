package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestBPlusTreeCRUDAndRange(t *testing.T) {
	pager, err := NewFilePageManager(filepath.Join(t.TempDir(), "btree.db"))
	if err != nil {
		t.Fatalf("NewFilePageManager() error = %v", err)
	}
	defer func() {
		_ = pager.Close()
	}()

	tree := NewBPlusTree(16, pager)
	ctx := context.Background()
	if err = tree.Insert(ctx, []byte("a"), []byte("1")); err != nil {
		t.Fatalf("Insert(a) error = %v", err)
	}
	if err = tree.Insert(ctx, []byte("b"), []byte("2")); err != nil {
		t.Fatalf("Insert(b) error = %v", err)
	}
	if err = tree.Insert(ctx, []byte("c"), []byte("3")); err != nil {
		t.Fatalf("Insert(c) error = %v", err)
	}

	value, err := tree.Search(ctx, []byte("b"))
	if err != nil || string(value) != "2" {
		t.Fatalf("Search(b) = (%q, %v), want (2, nil)", string(value), err)
	}

	it, err := tree.RangeScan(ctx, []byte("a"), []byte("b"))
	if err != nil {
		t.Fatalf("RangeScan() error = %v", err)
	}
	count := 0
	for it.Next() {
		count++
	}
	if count != 2 {
		t.Fatalf("RangeScan count = %d, want 2", count)
	}

	if err = tree.Delete(ctx, []byte("b")); err != nil {
		t.Fatalf("Delete(b) error = %v", err)
	}
	if _, err = tree.Search(ctx, []byte("b")); err == nil {
		t.Fatal("Search(b) after delete should fail")
	}
}

func BenchmarkBPlusTreeInsert(b *testing.B) {
	tree := NewBPlusTree(16, nil)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		key := []byte{byte(i), byte(i >> 8), byte(i >> 16)}
		if err := tree.Insert(ctx, key, key); err != nil {
			b.Fatalf("Insert() error = %v", err)
		}
	}
}
