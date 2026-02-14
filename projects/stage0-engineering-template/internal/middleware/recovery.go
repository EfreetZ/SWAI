package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/EfreetZ/SWAI/projects/stage0-engineering-template/pkg/response"
)

// Recovery 中间件：捕获 panic，防止单个请求崩溃导致整个服务宕机
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// 记录 panic 堆栈信息
				slog.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"request_id", GetRequestID(r.Context()),
					"method", r.Method,
					"path", r.URL.Path,
				)
				// 返回 500 错误
				response.Error(w, http.StatusInternalServerError,
					response.CodeInternal, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
