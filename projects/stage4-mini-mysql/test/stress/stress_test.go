package stress

import (
	"context"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/executor"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/parser"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/wal"
)

// TestStressConcurrentInsert 并发写入压力测试。
func TestStressConcurrentInsert(t *testing.T) {
	pager, err := storage.NewFilePageManager(filepath.Join(t.TempDir(), "stress.db"))
	if err != nil {
		t.Fatalf("NewFilePageManager() error = %v", err)
	}
	defer func() { _ = pager.Close() }()
	walWriter, err := wal.NewWriter(filepath.Join(t.TempDir(), "stress.wal"))
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	defer func() { _ = walWriter.Close() }()

	engine := executor.NewWithDefaults(storage.NewBPlusTree(16, pager), walWriter)
	ctx := context.Background()

	const workers = 32
	const each = 200
	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < each; i++ {
				key := "k" + strconv.Itoa(workerID) + "-" + strconv.Itoa(i)
				_, _ = engine.Execute(ctx, "stress-session", &parser.InsertStmt{Table: "kv", Key: key, Value: "v"})
			}
		}(w)
	}
	wg.Wait()
}
