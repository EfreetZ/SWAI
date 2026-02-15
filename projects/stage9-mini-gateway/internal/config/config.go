package config

import "time"

// GatewayConfig 网关配置。
type GatewayConfig struct {
	Server ServerConfig
	Routes []RouteConfig
}

// ServerConfig 服务配置。
type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// RouteConfig 路由配置。
type RouteConfig struct {
	Name        string
	Prefix      string
	StripPrefix bool
	Methods     []string
	Targets     []TargetConfig
	APIKey      string
	QPS         int
	Burst       int
}

// TargetConfig 后端节点。
type TargetConfig struct {
	Addr   string
	Weight int
}
