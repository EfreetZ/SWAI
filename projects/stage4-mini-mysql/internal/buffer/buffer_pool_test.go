package buffer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
)

func TestBufferPoolFetchUnpinFlush(t *testing.T) {
	pager, err := storage.NewFilePageManager(filepath.Join(t.TempDir(), "buffer.db"))
	if err != nil {
		t.Fatalf("NewFilePageManager() error = %v", err)
	}
	defer func() {
		_ = pager.Close()
	}()

	bpm := NewBufferPoolManager(2, pager)
	ctx := context.Background()

	page, err := bpm.NewPage(ctx)
	if err != nil {
		t.Fatalf("NewPage() error = %v", err)
	}
	copy(page.Data[:], []byte("hello"))
	if err = bpm.UnpinPage(ctx, page.ID, true); err != nil {
		t.Fatalf("UnpinPage() error = %v", err)
	}
	if err = bpm.FlushPage(ctx, page.ID); err != nil {
		t.Fatalf("FlushPage() error = %v", err)
	}

	fetched, err := bpm.FetchPage(ctx, page.ID)
	if err != nil {
		t.Fatalf("FetchPage() error = %v", err)
	}
	if string(fetched.Data[:5]) != "hello" {
		t.Fatalf("fetched data = %q, want hello", string(fetched.Data[:5]))
	}
}

func TestLRUReplacer(t *testing.T) {
	r := NewLRUReplacer()
	r.Unpin(1)
	r.Unpin(2)
	r.Unpin(3)
	victim, ok := r.Victim()
	if !ok || victim != 1 {
		t.Fatalf("Victim() = (%d, %v), want (1, true)", victim, ok)
	}
}
