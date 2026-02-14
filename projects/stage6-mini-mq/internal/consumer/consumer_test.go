package consumer

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
)

func TestConsumerPollCommit(t *testing.T) {
	b := broker.NewBroker(1, "127.0.0.1:19092", t.TempDir())
	if err := b.CreateTopic("events", broker.TopicConfig{NumPartitions: 1, SegmentBytes: 1024}); err != nil {
		t.Fatalf("create topic failed: %v", err)
	}
	if _, err := b.Produce(context.Background(), "events", 0, []byte("k"), []byte("v")); err != nil {
		t.Fatalf("produce failed: %v", err)
	}

	c := NewClient(b, "g1", "m1")
	msg, err := c.Poll(context.Background(), "events", 0)
	if err != nil {
		t.Fatalf("poll failed: %v", err)
	}
	if string(msg.Value) != "v" {
		t.Fatalf("unexpected value: %s", string(msg.Value))
	}
	c.Commit("events", 0)
	if got := b.OffsetMgr.Get("g1", "events", 0); got != 1 {
		t.Fatalf("unexpected committed offset: %d", got)
	}
}
