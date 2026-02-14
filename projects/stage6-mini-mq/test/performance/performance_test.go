package performance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/producer"
)

func TestProduceThroughput(t *testing.T) {
	mq := broker.NewBroker(1, "127.0.0.1:19092", t.TempDir())
	if err := mq.CreateTopic("perf", broker.TopicConfig{NumPartitions: 4, SegmentBytes: 1 << 20}); err != nil {
		t.Fatalf("create topic failed: %v", err)
	}
	p := producer.NewClient(mq, 200, &producer.KeyHashPartitioner{})
	ctx := context.Background()

	start := time.Now()
	const total = 5000
	for i := 0; i < total; i++ {
		if err := p.Send(ctx, "perf", []byte(fmt.Sprintf("k-%d", i)), []byte("payload")); err != nil {
			t.Fatalf("send failed: %v", err)
		}
	}
	elapsed := time.Since(start)
	if elapsed > 3*time.Second {
		t.Fatalf("performance regression, elapsed=%s", elapsed)
	}
}
