// Package logger 提供结构化日志能力
// 基于 Go 1.21+ 标准库 slog，支持 JSON 和 Text 两种输出格式
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New 根据配置创建 slog.Logger 实例并设为全局默认
// level: debug/info/warn/error
// format: json/text
func New(level, format string) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	switch strings.ToLower(format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// parseLevel 将字符串日志级别转换为 slog.Level
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
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
