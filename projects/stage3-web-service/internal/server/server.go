package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// HTTPServer 封装 HTTP 服务。
type HTTPServer struct {
	server *http.Server
	logger *slog.Logger
}

// NewHTTPServer 创建 HTTP 服务。
func NewHTTPServer(port int, handler http.Handler, logger *slog.Logger) *HTTPServer {
	return &HTTPServer{
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
		logger: logger,
	}
}

// Start 启动 HTTP 服务。
func (s *HTTPServer) Start() error {
	s.logger.Info("http server starting", "addr", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown 优雅关闭 HTTP 服务。
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
