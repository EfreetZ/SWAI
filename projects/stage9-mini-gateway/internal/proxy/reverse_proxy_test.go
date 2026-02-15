package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestForward(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer backend.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	w := httptest.NewRecorder()
	if err := Forward(w, req, backend.URL, "/api"); err != nil {
		t.Fatalf("forward failed: %v", err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}
