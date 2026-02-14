package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthCheck 测试健康检查端点返回正确数据
func TestHealthCheck(t *testing.T) {
	h := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Check(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Code    int `json:"code"`
		Data    struct {
			Status  string `json:"status"`
			Version string `json:"version"`
		} `json:"data"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Code != 0 {
		t.Errorf("code = %d, want 0", resp.Code)
	}
	if resp.Data.Status != "ok" {
		t.Errorf("status = %q, want %q", resp.Data.Status, "ok")
	}
	if resp.Message != "success" {
		t.Errorf("message = %q, want %q", resp.Message, "success")
	}
}

// TestHealthCheckMethodNotAllowed 测试非 GET 方法返回 405
func TestHealthCheckMethodNotAllowed(t *testing.T) {
	h := NewHealthHandler()
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/health", nil)
		rec := httptest.NewRecorder()
		h.Check(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s /health status = %d, want %d", method, rec.Code, http.StatusMethodNotAllowed)
		}
	}
}

// BenchmarkHealthCheck 基准测试：健康检查接口性能
func BenchmarkHealthCheck(b *testing.B) {
	h := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		h.Check(rec, req)
	}
}
