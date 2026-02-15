package chaos

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/config"
	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/gateway"
)

func TestBackendDownChaos(t *testing.T) {
	cfg := config.GatewayConfig{Server: config.ServerConfig{Addr: ":0", ReadTimeout: time.Second, WriteTimeout: time.Second}, Routes: []config.RouteConfig{{Name: "demo", Prefix: "/api", Methods: []string{"GET"}, Targets: []config.TargetConfig{{Addr: "http://127.0.0.1:1", Weight: 1}}, APIKey: "k", QPS: 100000, Burst: 100000}}}
	gw := gateway.NewGateway(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	req.Header.Set("X-API-Key", "k")
	w := httptest.NewRecorder()
	gw.Handler(w, req)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected bad gateway, got %d", w.Code)
	}
}
