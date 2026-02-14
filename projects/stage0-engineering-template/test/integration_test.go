// Package test 包含集成测试
// 集成测试验证多个模块协同工作的正确性
package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage0-engineering-template/internal/handler"
	"github.com/EfreetZ/SWAI/projects/stage0-engineering-template/internal/middleware"
)

// TestFullMiddlewareChain 集成测试：完整中间件链 + 健康检查
func TestFullMiddlewareChain(t *testing.T) {
	// 组装路由
	mux := http.NewServeMux()
	healthHandler := handler.NewHealthHandler()
	mux.HandleFunc("/health", healthHandler.Check)

	// 组装中间件链
	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.RequestID(h)
	h = middleware.Recovery(h)

	// 发送请求
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// 验证状态码
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// 验证 request_id 响应头存在
	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("missing X-Request-ID header")
	}

	// 验证响应 JSON 结构
	var resp struct {
		Code int    `json:"code"`
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Data.Status != "ok" {
		t.Errorf("status = %q, want %q", resp.Data.Status, "ok")
	}
}

// TestPanicRecoveryIntegration 集成测试：panic 被完整中间件链捕获
func TestPanicRecoveryIntegration(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("integration test panic")
	})

	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.RequestID(h)
	h = middleware.Recovery(h)

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	// 即使 panic，request_id 也应存在
	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("missing X-Request-ID header after panic")
	}
}

// Test404Integration 集成测试：访问不存在的路由
func Test404Integration(t *testing.T) {
	mux := http.NewServeMux()
	healthHandler := handler.NewHealthHandler()
	mux.HandleFunc("/health", healthHandler.Check)

	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.RequestID(h)
	h = middleware.Recovery(h)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
