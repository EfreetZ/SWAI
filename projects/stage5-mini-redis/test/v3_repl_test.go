package test

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/replication"
)

func TestV3Replication(t *testing.T) {
	master := replication.NewMaster()
	master.RegisterSlave("slave-1")
	master.RegisterSlave("slave-2")

	offset := master.Broadcast("SET k v")
	entries := master.PullSince(offset - 1)
	if len(entries) == 0 {
		t.Fatal("PullSince() should return replicated entries")
	}

	slave := replication.NewSlave("127.0.0.1", 16379)
	slave.Ack(offset)
	if slave.Offset() != offset {
		t.Fatalf("slave offset = %d, want %d", slave.Offset(), offset)
	}
}
