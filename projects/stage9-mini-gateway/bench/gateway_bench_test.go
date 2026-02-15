package bench

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/config"
	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/gateway"
)

func BenchmarkGateway(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer backend.Close()
	cfg := config.GatewayConfig{Server: config.ServerConfig{Addr: ":0", ReadTimeout: time.Second, WriteTimeout: time.Second}, Routes: []config.RouteConfig{{Name: "demo", Prefix: "/api", Methods: []string{"GET"}, Targets: []config.TargetConfig{{Addr: backend.URL, Weight: 1}}, APIKey: "k", QPS: 100000, Burst: 100000}}}
	gw := gateway.NewGateway(cfg)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
		req.Header.Set("X-API-Key", "k")
		w := httptest.NewRecorder()
		gw.Handler(w, req)
	}
}
