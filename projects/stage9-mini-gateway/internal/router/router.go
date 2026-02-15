package router

import (
	"net/http"
	"strings"

	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/config"
)

// Route 路由。
type Route struct {
	Config config.RouteConfig
}

// Router 前缀路由器。
type Router struct {
	routes []Route
}

// NewRouter 创建路由器。
func NewRouter(cfgs []config.RouteConfig) *Router {
	routes := make([]Route, 0, len(cfgs))
	for _, c := range cfgs {
		routes = append(routes, Route{Config: c})
	}
	return &Router{routes: routes}
}

// Match 匹配路由。
func (r *Router) Match(req *http.Request) (*Route, bool) {
	for i := range r.routes {
		route := &r.routes[i]
		if !strings.HasPrefix(req.URL.Path, route.Config.Prefix) {
			continue
		}
		if len(route.Config.Methods) == 0 {
			return route, true
		}
		for _, m := range route.Config.Methods {
			if strings.EqualFold(m, req.Method) {
				return route, true
			}
		}
	}
	return nil, false
}
