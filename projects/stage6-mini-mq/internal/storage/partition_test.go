package storage

import (
	"context"
	"testing"
)

func TestPartitionAppendRead(t *testing.T) {
	p, err := NewPartition("topic-a", 0, t.TempDir(), 128)
	if err != nil {
		t.Fatalf("new partition failed: %v", err)
	}
	t.Cleanup(func() { _ = p.Close() })

	offset, err := p.Append(context.Background(), []byte("k"), []byte("v"))
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}
	msg, err := p.Read(context.Background(), offset)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(msg.Value) != "v" {
		t.Fatalf("unexpected value: %s", string(msg.Value))
	}
}
