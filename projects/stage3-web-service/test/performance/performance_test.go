package performance

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	stage3test "github.com/EfreetZ/SWAI/projects/stage3-web-service/test"
)

// TestLoginPerformance 验证登录接口在固定样本下的吞吐表现。
func TestLoginPerformance(t *testing.T) {
	app, _ := stage3test.NewTestAppForExternal()

	registerBody := map[string]string{"username": "bench", "email": "bench@example.com", "password": "123456"}
	registerPayload, _ := json.Marshal(registerBody)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerPayload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	app.ServeHTTP(registerW, registerReq)

	for i := 0; i < 200; i++ {
		loginBody := map[string]string{"username": "bench", "password": "123456"}
		loginPayload, _ := json.Marshal(loginBody)
		loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginPayload))
		loginReq.Header.Set("Content-Type", "application/json")
		loginW := httptest.NewRecorder()
		app.ServeHTTP(loginW, loginReq)
		if loginW.Code != http.StatusOK {
			t.Fatalf("login status = %d, want %d", loginW.Code, http.StatusOK)
		}
	}
}
