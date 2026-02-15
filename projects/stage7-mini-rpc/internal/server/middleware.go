package server

import (
	"log/slog"
	"sync"
	"time"
)

// Context 服务端上下文。
type Context struct {
	Request   []byte
	Response  []byte
	Service   string
	Method    string
	Metadata  map[string]string
	StartTime time.Time
}

// HandlerFunc 处理函数。
type HandlerFunc func(ctx *Context) error

// Middleware 中间件函数。
type Middleware func(next HandlerFunc) HandlerFunc

// ChainMiddleware 构造中间件链。
func ChainMiddleware(final HandlerFunc, middlewares ...Middleware) HandlerFunc {
	h := final
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// LoggingMiddleware 日志中间件。
func LoggingMiddleware(logger *slog.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			start := time.Now()
			err := next(ctx)
			logger.Info("rpc request", "service", ctx.Service, "method", ctx.Method, "latency", time.Since(start), "error", err)
			return err
		}
	}
}

// RecoveryMiddleware panic 恢复中间件。
func RecoveryMiddleware(logger *slog.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("rpc panic recovered", "panic", r)
					err = ErrInternal
				}
			}()
			return next(ctx)
		}
	}
}

// RateLimitMiddleware 简单定频限流。
func RateLimitMiddleware(qps int) Middleware {
	if qps <= 0 {
		qps = 1000
	}
	interval := time.Second / time.Duration(qps)
	var mu sync.Mutex
	last := time.Now().Add(-interval)

	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) error {
			mu.Lock()
			now := time.Now()
			if now.Sub(last) < interval {
				mu.Unlock()
				return ErrRateLimited
			}
			last = now
			mu.Unlock()
			return next(ctx)
		}
	}
}
