package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthRegisterLoginRefresh(t *testing.T) {
	app, _ := newTestApp()

	registerBody := map[string]string{"username": "alice", "email": "alice@example.com", "password": "123456"}
	registerPayload, _ := json.Marshal(registerBody)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerPayload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	app.ServeHTTP(registerW, registerReq)
	if registerW.Code != http.StatusOK {
		t.Fatalf("register status = %d, want %d", registerW.Code, http.StatusOK)
	}

	loginBody := map[string]string{"username": "alice", "password": "123456"}
	loginPayload, _ := json.Marshal(loginBody)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	app.ServeHTTP(loginW, loginReq)
	if loginW.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d", loginW.Code, http.StatusOK)
	}

	var loginResp struct {
		Code int `json:"code"`
		Data struct {
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("unmarshal login response error = %v", err)
	}
	if loginResp.Data.RefreshToken == "" {
		t.Fatal("refresh token should not be empty")
	}

	refreshBody := map[string]string{"refresh_token": loginResp.Data.RefreshToken}
	refreshPayload, _ := json.Marshal(refreshBody)
	refreshReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refreshPayload))
	refreshReq.Header.Set("Content-Type", "application/json")
	refreshW := httptest.NewRecorder()
	app.ServeHTTP(refreshW, refreshReq)
	if refreshW.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, want %d", refreshW.Code, http.StatusOK)
	}
}
