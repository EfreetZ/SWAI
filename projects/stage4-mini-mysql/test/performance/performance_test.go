package performance

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/executor"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/parser"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/wal"
)

// TestInsertPerformance 验证批量插入性能。
func TestInsertPerformance(t *testing.T) {
	pager, err := storage.NewFilePageManager(filepath.Join(t.TempDir(), "perf.db"))
	if err != nil {
		t.Fatalf("NewFilePageManager() error = %v", err)
	}
	defer func() { _ = pager.Close() }()
	walWriter, err := wal.NewWriter(filepath.Join(t.TempDir(), "perf.wal"))
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	defer func() { _ = walWriter.Close() }()

	engine := executor.NewWithDefaults(storage.NewBPlusTree(16, pager), walWriter)
	ctx := context.Background()
	start := time.Now()
	for i := 0; i < 5000; i++ {
		stmt := &parser.InsertStmt{Table: "kv", Key: string(rune(i % 20000)), Value: "v"}
		if _, err = engine.Execute(ctx, "perf-session", stmt); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	}
	t.Logf("insert 5000 elapsed=%s", time.Since(start))
}
