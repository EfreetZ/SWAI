package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(io.Writer(os.Stdout), nil))
	b := broker.NewBroker(1, "127.0.0.1:19092", ".")
	_ = b.CreateTopic("events", broker.TopicConfig{NumPartitions: 3, SegmentBytes: 1024 * 1024})

	handler := server.NewHandler(b)
	srv := server.NewTCPServer("127.0.0.1:19092", handler, logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := srv.Start(ctx); err != nil {
		logger.Error("broker stopped", "error", err)
		os.Exit(1)
	}
}
