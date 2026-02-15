package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/server"
)

func TestIntegration(t *testing.T) {
	s := server.NewServer()
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/ping")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "pong" {
		t.Fatalf("unexpected body: %s", string(body))
	}
}
