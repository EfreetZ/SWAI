package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerRoutes(t *testing.T) {
	s := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}
