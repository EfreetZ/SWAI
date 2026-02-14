package chaos

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/executor"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/parser"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/wal"
)

// TestChaosRollback 模拟事务回滚场景，验证未提交写入不会生效。
func TestChaosRollback(t *testing.T) {
	pager, err := storage.NewFilePageManager(filepath.Join(t.TempDir(), "chaos.db"))
	if err != nil {
		t.Fatalf("NewFilePageManager() error = %v", err)
	}
	defer func() { _ = pager.Close() }()
	walWriter, err := wal.NewWriter(filepath.Join(t.TempDir(), "chaos.wal"))
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	defer func() { _ = walWriter.Close() }()

	engine := executor.NewWithDefaults(storage.NewBPlusTree(16, pager), walWriter)
	ctx := context.Background()
	sessionID := "chaos-session"

	if _, err = engine.Execute(ctx, sessionID, &parser.TxStmt{Action: "BEGIN"}); err != nil {
		t.Fatalf("BEGIN error = %v", err)
	}
	if _, err = engine.Execute(ctx, sessionID, &parser.InsertStmt{Table: "kv", Key: "k1", Value: "v1"}); err != nil {
		t.Fatalf("INSERT error = %v", err)
	}
	if _, err = engine.Execute(ctx, sessionID, &parser.TxStmt{Action: "ROLLBACK"}); err != nil {
		t.Fatalf("ROLLBACK error = %v", err)
	}

	result, err := engine.Execute(ctx, sessionID, &parser.SelectStmt{Table: "kv", Key: "k1"})
	if err != nil {
		t.Fatalf("SELECT error = %v", err)
	}
	if result != "NULL" {
		t.Fatalf("result = %q, want NULL", result)
	}
}
