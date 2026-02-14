package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/consumer"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/producer"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	b := broker.NewBroker(1, "127.0.0.1:19092", ".")
	_ = b.CreateTopic("events", broker.TopicConfig{NumPartitions: 3, SegmentBytes: 1024 * 1024})

	prod := producer.NewClient(b, 10, &producer.RoundRobinPartitioner{})
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_ = prod.Send(ctx, "events", []byte(fmt.Sprintf("k-%d", i)), []byte(fmt.Sprintf("v-%d", i)))
	}

	cons := consumer.NewClient(b, "group-1", "member-1")
	pollCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	for i := 0; i < 2; i++ {
		msg, err := cons.Poll(pollCtx, "events", i)
		if err != nil {
			logger.Error("poll failed", "error", err)
			continue
		}
		logger.Info("consume message", "offset", msg.Offset, "key", string(msg.Key), "value", string(msg.Value))
		cons.Commit("events", i)
	}
}
