package test

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/handler"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/pkg"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/repository"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/server"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/service"
)

func newTestApp() (http.Handler, *pkg.JWTManager) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	jwtMgr := pkg.NewJWTManager("test-secret", time.Minute, time.Hour)
	userRepo := repository.NewInMemoryUserRepo()
	roleRepo := repository.NewInMemoryRoleRepo()
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	userSvc := service.NewUserService(userRepo)
	roleSvc := service.NewRoleService(roleRepo)
	rbacChecker := service.NewRBACChecker(roleRepo)

	router := server.NewRouter(server.Dependencies{
		Logger:      logger,
		JWTManager:  jwtMgr,
		RBACChecker: rbacChecker,
		AuthHandler: handler.NewAuthHandler(authSvc),
		UserHandler: handler.NewUserHandler(userSvc),
		RoleHandler: handler.NewRoleHandler(roleSvc),
	})
	return router, jwtMgr
}

// NewTestAppForExternal 提供给子测试包复用测试应用实例。
func NewTestAppForExternal() (http.Handler, *pkg.JWTManager) {
	return newTestApp()
}
