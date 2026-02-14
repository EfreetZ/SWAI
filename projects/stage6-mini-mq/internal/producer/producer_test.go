package producer

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
)

func TestProducerSend(t *testing.T) {
	b := broker.NewBroker(1, "127.0.0.1:19092", t.TempDir())
	if err := b.CreateTopic("events", broker.TopicConfig{NumPartitions: 2, SegmentBytes: 1024}); err != nil {
		t.Fatalf("create topic failed: %v", err)
	}
	p := NewClient(b, 2, &KeyHashPartitioner{})
	if err := p.Send(context.Background(), "events", []byte("k1"), []byte("v1")); err != nil {
		t.Fatalf("send failed: %v", err)
	}
}
