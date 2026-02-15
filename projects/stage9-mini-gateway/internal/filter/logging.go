package filter

import (
	"log/slog"
	"net/http"
	"time"
)

// Logging 日志过滤器。
func Logging(logger *slog.Logger) Filter {
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next(w, r)
			logger.Info("gateway request", "method", r.Method, "path", r.URL.Path, "latency", time.Since(start))
		}
	}
}
