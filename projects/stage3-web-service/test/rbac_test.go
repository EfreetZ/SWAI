package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRBACDenyViewer(t *testing.T) {
	app, jwtMgr := newTestApp()
	access, _, err := jwtMgr.GenerateTokenPair(1, "viewer-user", "viewer")
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}
