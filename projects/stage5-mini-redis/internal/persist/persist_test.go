package persist

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
)

func TestAOFAppendReplay(t *testing.T) {
	aof, err := NewAOF(filepath.Join(t.TempDir(), "test.aof"))
	if err != nil {
		t.Fatalf("NewAOF() error = %v", err)
	}
	defer func() { _ = aof.Close() }()

	ctx := context.Background()
	_ = aof.Append(ctx, []string{"SET", "k", "v"})
	_ = aof.Flush(ctx)
	commands, err := aof.Replay(ctx)
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	if len(commands) != 1 || commands[0][0] != "SET" {
		t.Fatalf("commands = %v, want one SET", commands)
	}
}

func TestRDBSnapshot(t *testing.T) {
	d := db.New()
	ctx := context.Background()
	_ = d.SetString(ctx, "k", "v", 0)

	path := filepath.Join(t.TempDir(), "test.rdb")
	snapshot := d.Snapshot()
	if err := SaveSnapshot(ctx, path, snapshot); err != nil {
		t.Fatalf("SaveSnapshot() error = %v", err)
	}
	loaded, err := LoadSnapshot(ctx, path)
	if err != nil {
		t.Fatalf("LoadSnapshot() error = %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("len(loaded) = %d, want 1", len(loaded))
	}
}
