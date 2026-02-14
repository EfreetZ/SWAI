package bench

import (
	"context"
	"fmt"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/producer"
)

func BenchmarkProduce(b *testing.B) {
	mq := broker.NewBroker(1, "127.0.0.1:19092", b.TempDir())
	if err := mq.CreateTopic("bench", broker.TopicConfig{NumPartitions: 4, SegmentBytes: 1 << 20}); err != nil {
		b.Fatalf("create topic failed: %v", err)
	}
	p := producer.NewClient(mq, 100, &producer.KeyHashPartitioner{})
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("k-%d", i))
		if err := p.Send(ctx, "bench", key, []byte("value")); err != nil {
			b.Fatalf("send failed: %v", err)
		}
	}
}
