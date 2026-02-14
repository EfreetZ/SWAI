package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestSegmentAppendRead(t *testing.T) {
	dir := t.TempDir()
	seg, err := NewSegment(filepath.Join(dir, "p0"), 0, 1024*1024)
	if err != nil {
		t.Fatalf("new segment failed: %v", err)
	}
	t.Cleanup(func() { _ = seg.Close() })

	offset, _, err := seg.Append(context.Background(), &Message{Key: []byte("k1"), Value: []byte("v1")})
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}
	if offset != 0 {
		t.Fatalf("unexpected offset: %d", offset)
	}

	msg, err := seg.Read(context.Background(), 0)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(msg.Key) != "k1" || string(msg.Value) != "v1" {
		t.Fatalf("unexpected message: key=%s value=%s", string(msg.Key), string(msg.Value))
	}
}

func TestSegmentReadNotFound(t *testing.T) {
	dir := t.TempDir()
	seg, err := NewSegment(filepath.Join(dir, "p0"), 0, 1024*1024)
	if err != nil {
		t.Fatalf("new segment failed: %v", err)
	}
	t.Cleanup(func() { _ = seg.Close() })

	_, err = seg.Read(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error")
	}
}
