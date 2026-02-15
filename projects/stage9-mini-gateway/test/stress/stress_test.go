package stress

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/config"
	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/gateway"
)

func TestGatewayStress(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer backend.Close()

	cfg := config.GatewayConfig{Server: config.ServerConfig{Addr: ":0", ReadTimeout: time.Second, WriteTimeout: time.Second}, Routes: []config.RouteConfig{{Name: "demo", Prefix: "/api", Methods: []string{"GET"}, Targets: []config.TargetConfig{{Addr: backend.URL, Weight: 1}}, APIKey: "k", QPS: 100000, Burst: 100000}}}
	gw := gateway.NewGateway(cfg)

	const workers = 20
	const each = 300
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < each; i++ {
				req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
				req.Header.Set("X-API-Key", "k")
				resp := httptest.NewRecorder()
				gw.Handler(resp, req)
				if resp.Code != http.StatusOK {
					t.Errorf("unexpected status: %d", resp.Code)
					return
				}
			}
		}()
	}
	wg.Wait()
}
