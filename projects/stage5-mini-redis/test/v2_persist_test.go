package test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/persist"
)

func TestV2Persist(t *testing.T) {
	d := db.New()
	ctx := context.Background()
	_, _ = d.ExecuteCommand(ctx, []string{"SET", "k", "v"})

	aof, err := persist.NewAOF(filepath.Join(t.TempDir(), "persist.aof"))
	if err != nil {
		t.Fatalf("NewAOF() error = %v", err)
	}
	defer func() { _ = aof.Close() }()

	_ = aof.Append(ctx, []string{"SET", "k", "v"})
	_ = aof.Flush(ctx)
	commands, err := aof.Replay(ctx)
	if err != nil || len(commands) != 1 {
		t.Fatalf("Replay() = (%v, %v)", commands, err)
	}

	snapshotPath := filepath.Join(t.TempDir(), "persist.rdb")
	if err = persist.SaveSnapshot(ctx, snapshotPath, d.Snapshot()); err != nil {
		t.Fatalf("SaveSnapshot() error = %v", err)
	}
	loaded, err := persist.LoadSnapshot(ctx, snapshotPath)
	if err != nil || len(loaded) != 1 {
		t.Fatalf("LoadSnapshot() = (%v, %v)", loaded, err)
	}
}
