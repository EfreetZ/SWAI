package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/config"
)

func TestGatewayHandler(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	}))
	defer backend.Close()

	cfg := config.GatewayConfig{Server: config.ServerConfig{Addr: ":0", ReadTimeout: time.Second, WriteTimeout: time.Second}, Routes: []config.RouteConfig{{Name: "demo", Prefix: "/api", Methods: []string{"GET"}, Targets: []config.TargetConfig{{Addr: backend.URL, Weight: 1}}, APIKey: "k", QPS: 1000, Burst: 1000}}}
	gw := NewGateway(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	req.Header.Set("X-API-Key", "k")
	w := httptest.NewRecorder()
	gw.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	if w.Body.String() != "pong" {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}
