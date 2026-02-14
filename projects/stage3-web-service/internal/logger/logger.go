package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New 创建结构化日志实例。
func New(level, format string) *slog.Logger {
	handlerOptions := &slog.HandlerOptions{Level: parseLevel(level)}
	if strings.EqualFold(format, "json") {
		return slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
