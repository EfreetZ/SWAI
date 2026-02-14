// Package middleware 提供 HTTP 中间件
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// 用于 context 中存取 request_id 的 key
type contextKey string

const requestIDKey contextKey = "request_id"

// RequestID 中间件：为每个请求生成唯一 ID，写入 context 和响应头
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 优先从请求头获取（上游已生成），否则自动生成
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = generateID()
		}

		// 写入响应头，便于客户端和链路追踪关联
		w.Header().Set("X-Request-ID", id)

		// 写入 context，后续 handler/service 可通过 GetRequestID 获取
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID 从 context 中提取 request_id
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// generateID 生成 16 字节的随机十六进制字符串
func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
