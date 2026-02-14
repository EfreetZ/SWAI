package test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/executor"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/parser"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/wal"
)

func newEngine(t *testing.T) *executor.Engine {
	t.Helper()
	pager, err := storage.NewFilePageManager(filepath.Join(t.TempDir(), "integration.db"))
	if err != nil {
		t.Fatalf("NewFilePageManager() error = %v", err)
	}
	t.Cleanup(func() { _ = pager.Close() })

	walWriter, err := wal.NewWriter(filepath.Join(t.TempDir(), "integration.wal"))
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	t.Cleanup(func() { _ = walWriter.Close() })

	tree := storage.NewBPlusTree(16, pager)
	return executor.NewWithDefaults(tree, walWriter)
}

func TestTransactionFlow(t *testing.T) {
	engine := newEngine(t)
	ctx := context.Background()
	sessionID := "integration-session"

	if _, err := engine.Execute(ctx, sessionID, &parser.TxStmt{Action: "BEGIN"}); err != nil {
		t.Fatalf("BEGIN error = %v", err)
	}
	if _, err := engine.Execute(ctx, sessionID, &parser.InsertStmt{Table: "kv", Key: "k1", Value: "v1"}); err != nil {
		t.Fatalf("INSERT error = %v", err)
	}
	if _, err := engine.Execute(ctx, sessionID, &parser.TxStmt{Action: "COMMIT"}); err != nil {
		t.Fatalf("COMMIT error = %v", err)
	}

	result, err := engine.Execute(ctx, sessionID, &parser.SelectStmt{Table: "kv", Key: "k1"})
	if err != nil {
		t.Fatalf("SELECT error = %v", err)
	}
	if result != "v1" {
		t.Fatalf("SELECT result = %q, want v1", result)
	}
}
