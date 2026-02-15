package main

import (
	"log"
	"net/http"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/config"
	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/gateway"
)

func main() {
	cfg := config.GatewayConfig{Server: config.ServerConfig{Addr: ":18090", ReadTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second}, Routes: []config.RouteConfig{{Name: "demo", Prefix: "/api", StripPrefix: true, Methods: []string{"GET", "POST"}, Targets: []config.TargetConfig{{Addr: "http://127.0.0.1:18091", Weight: 1}}, APIKey: "demo-key", QPS: 2000, Burst: 3000}}}
	gw := gateway.NewGateway(cfg)
	server := &http.Server{Addr: cfg.Server.Addr, Handler: http.HandlerFunc(gw.Handler), ReadTimeout: cfg.Server.ReadTimeout, WriteTimeout: cfg.Server.WriteTimeout}
	log.Fatal(server.ListenAndServe())
}
