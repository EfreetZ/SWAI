package server

import (
	"log/slog"
	"net/http"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/handler"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/middleware"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/pkg"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/service"
)

// Dependencies 组装路由依赖。
type Dependencies struct {
	Logger      *slog.Logger
	JWTManager  *pkg.JWTManager
	RBACChecker *service.RBACChecker
	AuthHandler *handler.AuthHandler
	UserHandler *handler.UserHandler
	RoleHandler *handler.RoleHandler
}

// NewRouter 创建 HTTP 路由。
func NewRouter(dep Dependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		handler.Success(w, map[string]any{"status": "ok", "service": "stage3-web-service"})
	})

	mux.HandleFunc("/api/v1/auth/register", dep.AuthHandler.Register)
	mux.HandleFunc("/api/v1/auth/login", dep.AuthHandler.Login)
	mux.HandleFunc("/api/v1/auth/refresh", dep.AuthHandler.Refresh)

	protectedMux := http.NewServeMux()

	userListChain := chain(
		http.HandlerFunc(dep.UserHandler.ListUsers),
		func(next http.Handler) http.Handler { return middleware.RBAC(dep.RBACChecker, "user", "read", next) },
	)
	userByIDChain := chain(
		http.HandlerFunc(dep.UserHandler.UserByID),
		func(next http.Handler) http.Handler { return middleware.RBAC(dep.RBACChecker, "user", "read", next) },
	)
	roleListChain := chain(
		http.HandlerFunc(dep.RoleHandler.ListRoles),
		func(next http.Handler) http.Handler { return middleware.RBAC(dep.RBACChecker, "role", "read", next) },
	)
	protectedMux.Handle("/api/v1/users", userListChain)
	protectedMux.Handle("/api/v1/users/", userByIDChain)
	protectedMux.Handle("/api/v1/roles", roleListChain)

	protected := middleware.Auth(dep.JWTManager, protectedMux)
	mux.Handle("/api/v1/users", protected)
	mux.Handle("/api/v1/users/", protected)
	mux.Handle("/api/v1/roles", protected)

	return chain(mux,
		middleware.RequestID,
		func(next http.Handler) http.Handler { return middleware.CORS(next) },
		func(next http.Handler) http.Handler { return middleware.Recovery(dep.Logger, next) },
		func(next http.Handler) http.Handler { return middleware.Logger(dep.Logger, next) },
	)
}

func chain(base http.Handler, wrappers ...func(http.Handler) http.Handler) http.Handler {
	h := base
	for i := len(wrappers) - 1; i >= 0; i-- {
		h = wrappers[i](h)
	}
	return h
}
