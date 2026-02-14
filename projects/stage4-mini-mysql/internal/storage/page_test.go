package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestFilePageManagerReadWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	pager, err := NewFilePageManager(path)
	if err != nil {
		t.Fatalf("NewFilePageManager() error = %v", err)
	}
	defer func() {
		_ = pager.Close()
	}()

	pageID, err := pager.AllocatePage(context.Background())
	if err != nil {
		t.Fatalf("AllocatePage() error = %v", err)
	}
	page := &Page{ID: pageID}
	copy(page.Data[:], []byte("hello"))
	if err = pager.WritePage(context.Background(), page); err != nil {
		t.Fatalf("WritePage() error = %v", err)
	}

	readPage, err := pager.ReadPage(context.Background(), pageID)
	if err != nil {
		t.Fatalf("ReadPage() error = %v", err)
	}
	if string(readPage.Data[:5]) != "hello" {
		t.Fatalf("read data = %q, want hello", string(readPage.Data[:5]))
	}
}

func BenchmarkPageWrite(b *testing.B) {
	path := filepath.Join(b.TempDir(), "bench.db")
	pager, err := NewFilePageManager(path)
	if err != nil {
		b.Fatalf("NewFilePageManager() error = %v", err)
	}
	defer func() {
		_ = pager.Close()
	}()

	for i := 0; i < b.N; i++ {
		pageID, allocErr := pager.AllocatePage(context.Background())
		if allocErr != nil {
			b.Fatalf("AllocatePage() error = %v", allocErr)
		}
		page := &Page{ID: pageID}
		if writeErr := pager.WritePage(context.Background(), page); writeErr != nil {
			b.Fatalf("WritePage() error = %v", writeErr)
		}
	}
}
