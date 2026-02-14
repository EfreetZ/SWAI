package chaos

import (
	"context"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/producer"
)

func TestCanceledContextChaos(t *testing.T) {
	mq := broker.NewBroker(1, "127.0.0.1:19092", t.TempDir())
	if err := mq.CreateTopic("chaos", broker.TopicConfig{NumPartitions: 1, SegmentBytes: 1 << 20}); err != nil {
		t.Fatalf("create topic failed: %v", err)
	}
	p := producer.NewClient(mq, 10, &producer.RoundRobinPartitioner{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.Send(ctx, "chaos", []byte("k"), []byte("v")); err == nil {
		t.Fatal("expected canceled context error")
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel2()
	time.Sleep(time.Millisecond)
	if err := p.Send(ctx2, "chaos", []byte("k2"), []byte("v2")); err == nil {
		t.Fatal("expected deadline exceeded error")
	}
}
