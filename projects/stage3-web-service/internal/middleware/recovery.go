package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/handler"
)

// Recovery 捕获 panic 并返回统一错误响应。
func Recovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic recovered",
					"error", fmt.Sprintf("%v", rec),
					"stack", string(debug.Stack()),
					"request_id", RequestIDFromContext(r.Context()),
				)
				handler.Error(w, http.StatusInternalServerError, handler.ErrCodeInternal, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
