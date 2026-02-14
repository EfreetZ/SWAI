package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLogging 测试日志中间件正常执行且不影响响应
func TestLogging(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := Logging(inner)
	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "ok")
	}
}

// TestResponseWriterStatusCode 测试 responseWriter 正确记录状态码
func TestResponseWriterStatusCode(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	rw.WriteHeader(http.StatusNotFound)
	if rw.statusCode != http.StatusNotFound {
		t.Errorf("statusCode = %d, want %d", rw.statusCode, http.StatusNotFound)
	}
}

// TestResponseWriterBytes 测试 responseWriter 正确记录写入字节数
func TestResponseWriterBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	data := []byte("hello world")
	n, err := rw.Write(data)
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write() returned %d, want %d", n, len(data))
	}
	if rw.bytes != len(data) {
		t.Errorf("bytes = %d, want %d", rw.bytes, len(data))
	}
}

// BenchmarkLogging 基准测试：日志中间件性能开销
func BenchmarkLogging(b *testing.B) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Logging(inner)
	req := httptest.NewRequest(http.MethodGet, "/bench", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
