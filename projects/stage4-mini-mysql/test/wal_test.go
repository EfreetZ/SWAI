package test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/wal"
)

func TestWALRecoveryReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recover.wal")
	writer, err := wal.NewWriter(path)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	defer func() { _ = writer.Close() }()

	_, _ = writer.Append(context.Background(), &wal.LogRecord{TxID: 1, Type: wal.LogInsert, OldValue: []byte("k"), NewValue: []byte("v")})
	_ = writer.Flush(context.Background())

	tree := storage.NewBPlusTree(16, nil)
	recovery := wal.NewRecovery(writer, tree)
	if err = recovery.Replay(context.Background(), 1); err != nil {
		t.Fatalf("Replay() error = %v", err)
	}

	value, err := tree.Search(context.Background(), []byte("k"))
	if err != nil || string(value) != "v" {
		t.Fatalf("Search() = (%q, %v), want (v, nil)", string(value), err)
	}
}
