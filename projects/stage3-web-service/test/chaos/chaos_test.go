package chaos

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	stage3test "github.com/EfreetZ/SWAI/projects/stage3-web-service/test"
)

// TestChaosInvalidToken 模拟错误 token 与非法请求体，验证服务稳定返回标准错误。
func TestChaosInvalidToken(t *testing.T) {
	app, _ := stage3test.NewTestAppForExternal()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	badBody := bytes.NewBufferString("not-json")
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", badBody)
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	app.ServeHTTP(registerW, registerReq)
	if registerW.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", registerW.Code, http.StatusBadRequest)
	}

	_ = json.NewEncoder(bytes.NewBuffer(nil))
}
