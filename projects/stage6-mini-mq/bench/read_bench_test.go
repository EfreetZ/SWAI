package bench

import (
	"context"
	"fmt"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/consumer"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/producer"
)

func BenchmarkConsume(b *testing.B) {
	mq := broker.NewBroker(1, "127.0.0.1:19092", b.TempDir())
	if err := mq.CreateTopic("bench", broker.TopicConfig{NumPartitions: 1, SegmentBytes: 1 << 20}); err != nil {
		b.Fatalf("create topic failed: %v", err)
	}
	p := producer.NewClient(mq, 100, &producer.RoundRobinPartitioner{})
	ctx := context.Background()
	for i := 0; i < b.N+100; i++ {
		_ = p.Send(ctx, "bench", []byte(fmt.Sprintf("k-%d", i)), []byte("value"))
	}

	c := consumer.NewClient(mq, "g-bench", "m1")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.Poll(ctx, "bench", 0); err != nil {
			b.Fatalf("poll failed: %v", err)
		}
	}
}
