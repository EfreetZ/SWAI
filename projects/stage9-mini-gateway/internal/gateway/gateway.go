package gateway

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/balancer"
	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/breaker"
	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/config"
	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/filter"
	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/limiter"
	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/monitoring"
	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/proxy"
	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/router"
)

// Gateway 网关。
type Gateway struct {
	cfg      config.GatewayConfig
	router   *router.Router
	balancer balancer.Balancer
	breaker  *breaker.CircuitBreaker
	metrics  *monitoring.Metrics
	logger   *slog.Logger
}

// NewGateway 创建网关。
func NewGateway(cfg config.GatewayConfig) *Gateway {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return &Gateway{cfg: cfg, router: router.NewRouter(cfg.Routes), balancer: &balancer.RoundRobin{}, breaker: breaker.New(5, time.Second), metrics: &monitoring.Metrics{}, logger: logger}
}

// Handler 网关入口。
func (g *Gateway) Handler(w http.ResponseWriter, r *http.Request) {
	g.metrics.IncRequest()
	route, ok := g.router.Match(r)
	if !ok {
		g.metrics.IncError()
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("route not found"))
		return
	}

	bucket := limiter.NewTokenBucket(route.Config.QPS, route.Config.Burst)
	pipeline := filter.Chain(func(rw http.ResponseWriter, req *http.Request) {
		target, err := g.balancer.Pick(route.Config.Targets)
		if err != nil {
			g.metrics.IncError()
			rw.WriteHeader(http.StatusBadGateway)
			_, _ = rw.Write([]byte("no target"))
			return
		}
		err = g.breaker.Execute(func() error {
			return proxy.Forward(rw, req, target.Addr, route.Config.Prefix)
		})
		if err != nil {
			g.metrics.IncError()
			rw.WriteHeader(http.StatusBadGateway)
			_, _ = rw.Write([]byte(err.Error()))
			return
		}
	}, filter.CORS(), filter.APIKeyAuth(route.Config.APIKey), filter.RateLimit(bucket), filter.Logging(g.logger))

	pipeline(w, r)
}

// Metrics 获取指标。
func (g *Gateway) Metrics() *monitoring.Metrics {
	return g.metrics
}
