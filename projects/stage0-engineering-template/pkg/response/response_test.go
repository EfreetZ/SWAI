package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSuccess 测试成功响应格式
func TestSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	Success(rec, map[string]string{"key": "value"})

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Code != 0 {
		t.Errorf("code = %d, want 0", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("message = %q, want %q", resp.Message, "success")
	}

	// 验证 Content-Type
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json; charset=utf-8")
	}
}

// TestError 测试错误响应格式
func TestError(t *testing.T) {
	rec := httptest.NewRecorder()
	Error(rec, http.StatusBadRequest, CodeBadRequest, "invalid param")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Code != CodeBadRequest {
		t.Errorf("code = %d, want %d", resp.Code, CodeBadRequest)
	}
	if resp.Message != "invalid param" {
		t.Errorf("message = %q, want %q", resp.Message, "invalid param")
	}
	if resp.Data != nil {
		t.Errorf("data = %v, want nil", resp.Data)
	}
}

// BenchmarkSuccess 基准测试：成功响应序列化性能
func BenchmarkSuccess(b *testing.B) {
	data := map[string]string{"status": "ok"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		Success(rec, data)
	}
}
