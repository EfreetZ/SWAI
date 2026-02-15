package server

import (
	"io"
	"log/slog"
	"net/http"

	httpobs "github.com/EfreetZ/SWAI/projects/stage10-observability/internal/http"
	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/logging"
	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/metrics"
)

// Server 可观测性示例服务。
type Server struct {
	red    *metrics.REDMetrics
	logger *slog.Logger
	mux    *http.ServeMux
}

// NewServer 创建服务。
func NewServer() *Server {
	red := &metrics.REDMetrics{}
	logger := logging.NewLogger(io.Discard)
	mux := http.NewServeMux()
	s := &Server{red: red, logger: logger, mux: mux}
	h := httpobs.MetricsMiddleware(red, http.HandlerFunc(s.bizHandler))
	mux.Handle("/api/ping", h)
	mux.Handle("/metrics", httpobs.MetricsHandler(red))
	return s
}

// Handler 返回 http handler。
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) bizHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("biz request", "path", r.URL.Path)
	_, _ = w.Write([]byte("pong"))
}
