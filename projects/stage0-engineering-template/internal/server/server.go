// Package server 负责 HTTP 服务的启动与优雅关闭
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage0-engineering-template/internal/config"
	"github.com/EfreetZ/SWAI/projects/stage0-engineering-template/internal/handler"
	"github.com/EfreetZ/SWAI/projects/stage0-engineering-template/internal/middleware"
)

// Server HTTP 服务封装
type Server struct {
	httpServer *http.Server
	cfg        *config.Config
}

// New 创建 Server 实例，注册路由和中间件
func New(cfg *config.Config) *Server {
	mux := http.NewServeMux()

	// 注册路由
	healthHandler := handler.NewHealthHandler()
	mux.HandleFunc("/health", healthHandler.Check)

	// 中间件链：Recovery → RequestID → Logging → 路由
	// 执行顺序从外到内：请求先经过 Recovery，再 RequestID，再 Logging
	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.RequestID(h)
	h = middleware.Recovery(h)

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.App.Port),
			Handler:      h,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		cfg: cfg,
	}
}

// Run 启动 HTTP 服务并监听系统信号实现优雅关闭
// 优雅关闭流程：
// 1. 收到 SIGINT/SIGTERM 信号
// 2. 停止接受新连接
// 3. 等待已有连接处理完成（最多 30 秒）
// 4. 退出
func (s *Server) Run() error {
	// 启动 HTTP 服务（非阻塞）
	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting",
			"addr", s.httpServer.Addr,
			"env", s.cfg.App.Env,
		)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// 监听系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server failed to start: %w", err)
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig.String())
	}

	// 优雅关闭：给 30 秒处理剩余请求
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("server stopped gracefully")
	return nil
}
