package server

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// WaitForShutdown 等待信号并优雅关闭。
func WaitForShutdown(logger *slog.Logger, shutdown func(context.Context) error, timeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		return
	}
	logger.Info("server shutdown complete")
}
