package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/producer"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	b := broker.NewBroker(1, "127.0.0.1:19092", ".")
	_ = b.CreateTopic("events", broker.TopicConfig{NumPartitions: 3, SegmentBytes: 1024 * 1024})
	client := producer.NewClient(b, 100, &producer.RoundRobinPartitioner{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for i := 0; i < 10; i++ {
		key := []byte(fmt.Sprintf("k-%d", i))
		value := []byte(fmt.Sprintf("v-%d", i))
		if err := client.Send(ctx, "events", key, value); err != nil {
			logger.Error("send failed", "error", err)
			os.Exit(1)
		}
	}
	logger.Info("messages sent")
}
