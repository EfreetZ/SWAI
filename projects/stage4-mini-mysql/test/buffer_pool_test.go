package test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/buffer"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
)

func TestBufferPoolEviction(t *testing.T) {
	pager, err := storage.NewFilePageManager(filepath.Join(t.TempDir(), "buffer-evict.db"))
	if err != nil {
		t.Fatalf("NewFilePageManager() error = %v", err)
	}
	defer func() { _ = pager.Close() }()

	bpm := buffer.NewBufferPoolManager(1, pager)
	ctx := context.Background()

	p1, err := bpm.NewPage(ctx)
	if err != nil {
		t.Fatalf("NewPage() error = %v", err)
	}
	_ = bpm.UnpinPage(ctx, p1.ID, true)

	p2, err := bpm.NewPage(ctx)
	if err != nil {
		t.Fatalf("NewPage() second error = %v", err)
	}
	if p1.ID == p2.ID {
		t.Fatalf("expected different page ids")
	}
}
