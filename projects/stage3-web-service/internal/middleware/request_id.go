package middleware

import (
	"context"
	"net/http"
	"time"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
)

// RequestID 从请求头读取或生成请求 ID。
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext 获取请求 ID。
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if value, ok := ctx.Value(requestIDKey).(string); ok {
		return value
	}
	return ""
}

func generateRequestID() string {
	return time.Now().UTC().Format("20060102150405.000000000")
}
