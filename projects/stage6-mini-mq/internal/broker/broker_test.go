package broker

import (
	"context"
	"testing"
)

func TestBrokerProduceFetch(t *testing.T) {
	b := NewBroker(1, "127.0.0.1:19092", t.TempDir())
	if err := b.CreateTopic("events", TopicConfig{NumPartitions: 2, SegmentBytes: 1024}); err != nil {
		t.Fatalf("create topic failed: %v", err)
	}
	offset, err := b.Produce(context.Background(), "events", 0, []byte("k"), []byte("v"))
	if err != nil {
		t.Fatalf("produce failed: %v", err)
	}
	msg, err := b.Fetch(context.Background(), "events", 0, offset)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if string(msg.Value) != "v" {
		t.Fatalf("unexpected value: %s", string(msg.Value))
	}
}
