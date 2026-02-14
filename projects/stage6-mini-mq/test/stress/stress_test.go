package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/producer"
)

func TestConcurrentProduceStress(t *testing.T) {
	mq := broker.NewBroker(1, "127.0.0.1:19092", t.TempDir())
	if err := mq.CreateTopic("stress", broker.TopicConfig{NumPartitions: 8, SegmentBytes: 1 << 20}); err != nil {
		t.Fatalf("create topic failed: %v", err)
	}
	p := producer.NewClient(mq, 100, &producer.KeyHashPartitioner{})
	ctx := context.Background()

	const workers = 20
	const perWorker = 300
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		w := w
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				_ = p.Send(ctx, "stress", []byte(fmt.Sprintf("w%d-k%d", w, i)), []byte("v"))
			}
		}()
	}
	wg.Wait()
}
