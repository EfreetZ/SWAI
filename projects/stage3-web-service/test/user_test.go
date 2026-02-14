package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserEndpoints(t *testing.T) {
	app, jwtMgr := newTestApp()

	registerBody := map[string]string{"username": "admin", "email": "admin@example.com", "password": "123456"}
	registerPayload, _ := json.Marshal(registerBody)
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(registerPayload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	app.ServeHTTP(registerW, registerReq)
	if registerW.Code != http.StatusOK {
		t.Fatalf("register status = %d, want %d", registerW.Code, http.StatusOK)
	}

	access, _, err := jwtMgr.GenerateTokenPair(1, "admin", "admin")
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	listReq.Header.Set("Authorization", "Bearer "+access)
	listW := httptest.NewRecorder()
	app.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("list users status = %d, want %d", listW.Code, http.StatusOK)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/1", nil)
	getReq.Header.Set("Authorization", "Bearer "+access)
	getW := httptest.NewRecorder()
	app.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("get user status = %d, want %d", getW.Code, http.StatusOK)
	}
}
