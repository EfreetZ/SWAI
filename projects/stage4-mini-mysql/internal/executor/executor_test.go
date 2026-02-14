package executor

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/parser"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/wal"
)

func TestEngineExecute(t *testing.T) {
	pager, err := storage.NewFilePageManager(filepath.Join(t.TempDir(), "exec.db"))
	if err != nil {
		t.Fatalf("NewFilePageManager() error = %v", err)
	}
	defer func() {
		_ = pager.Close()
	}()

	walWriter, err := wal.NewWriter(filepath.Join(t.TempDir(), "exec.wal"))
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	defer func() {
		_ = walWriter.Close()
	}()

	tree := storage.NewBPlusTree(16, pager)
	engine := NewWithDefaults(tree, walWriter)
	ctx := context.Background()

	_, _ = engine.Execute(ctx, "s1", &parser.CreateTableStmt{Table: "kv"})
	if result, execErr := engine.Execute(ctx, "s1", &parser.InsertStmt{Table: "kv", Key: "a", Value: "1"}); execErr != nil || result != "OK" {
		t.Fatalf("insert result = (%q, %v)", result, execErr)
	}
	if result, execErr := engine.Execute(ctx, "s1", &parser.SelectStmt{Table: "kv", Key: "a"}); execErr != nil || result != "1" {
		t.Fatalf("select result = (%q, %v)", result, execErr)
	}
}
