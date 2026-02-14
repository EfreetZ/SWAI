package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/executor"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/server"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/wal"
)

func main() {
	logger := slog.New(slog.NewTextHandler(io.Writer(os.Stdout), nil))
	dataDir := "data"
	_ = os.MkdirAll(dataDir, 0o755)

	pageManager, err := storage.NewFilePageManager(filepath.Join(dataDir, "mini.db"))
	if err != nil {
		logger.Error("create page manager failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = pageManager.Close()
	}()

	walWriter, err := wal.NewWriter(filepath.Join(dataDir, "mini.wal"))
	if err != nil {
		logger.Error("create wal failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = walWriter.Close()
	}()

	tree := storage.NewBPlusTree(16, pageManager)
	engine := executor.NewWithDefaults(tree, walWriter)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	tcpServer := server.NewTCPServer("127.0.0.1:13306", engine, logger)
	if err = tcpServer.Start(ctx); err != nil {
		logger.Error("tcp server stopped with error", "error", err)
		os.Exit(1)
	}
}
