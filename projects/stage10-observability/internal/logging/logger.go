package logging

import (
	"io"
	"log/slog"
)

// NewLogger 创建结构化日志器。
func NewLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, nil))
}
