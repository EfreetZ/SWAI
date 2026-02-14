package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/server"
	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/ttl"
)

func main() {
	logger := slog.New(slog.NewTextHandler(io.Writer(os.Stdout), nil))
	database := db.New()
	ttlManager := ttl.NewManager(database)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	srv := server.NewTCPServer("127.0.0.1:16379", database, ttlManager, logger)
	if err := srv.Start(ctx); err != nil {
		logger.Error("mini-redis server stopped", "error", err)
		os.Exit(1)
	}
}
