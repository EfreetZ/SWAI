package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRequestIDGenerated 测试自动生成 request_id
func TestRequestIDGenerated(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id == "" {
			t.Error("expected request_id in context, got empty")
		}
		// 验证长度：16 字节 hex = 32 字符
		if len(id) != 32 {
			t.Errorf("request_id length = %d, want 32", len(id))
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// 响应头也应包含 request_id
	respID := rec.Header().Get("X-Request-ID")
	if respID == "" {
		t.Error("expected X-Request-ID in response header")
	}
}

// TestRequestIDFromHeader 测试从请求头传入 request_id
func TestRequestIDFromHeader(t *testing.T) {
	expectedID := "custom-trace-id-12345"

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id != expectedID {
			t.Errorf("request_id = %q, want %q", id, expectedID)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", expectedID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != expectedID {
		t.Errorf("response X-Request-ID = %q, want %q", got, expectedID)
	}
}

// TestGetRequestIDEmpty 测试空 context 返回空字符串
func TestGetRequestIDEmpty(t *testing.T) {
	id := GetRequestID(context.Background())
	if id != "" {
		t.Errorf("GetRequestID(empty ctx) = %q, want empty", id)
	}
}

// BenchmarkRequestID 基准测试：RequestID 中间件性能
func BenchmarkRequestID(b *testing.B) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
