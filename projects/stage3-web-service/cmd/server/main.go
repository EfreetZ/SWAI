package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/config"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/handler"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/logger"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/pkg"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/repository"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/server"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/service"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	appLogger := logger.New(cfg.Log.Level, cfg.Log.Format)
	jwtMgr := pkg.NewJWTManager(
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.AccessExpirySeconds)*time.Second,
		time.Duration(cfg.JWT.RefreshExpirySeconds)*time.Second,
	)

	userRepo := repository.NewInMemoryUserRepo()
	roleRepo := repository.NewInMemoryRoleRepo()
	authService := service.NewAuthService(userRepo, jwtMgr)
	userService := service.NewUserService(userRepo)
	roleService := service.NewRoleService(roleRepo)
	rbacChecker := service.NewRBACChecker(roleRepo)

	router := server.NewRouter(server.Dependencies{
		Logger:      appLogger,
		JWTManager:  jwtMgr,
		RBACChecker: rbacChecker,
		AuthHandler: handler.NewAuthHandler(authService),
		UserHandler: handler.NewUserHandler(userService),
		RoleHandler: handler.NewRoleHandler(roleService),
	})

	httpServer := server.NewHTTPServer(cfg.App.Port, router, appLogger)
	go func() {
		if serveErr := httpServer.Start(); serveErr != nil {
			appLogger.Error("server start failed", "error", serveErr)
			os.Exit(1)
		}
	}()

	server.WaitForShutdown(appLogger, httpServer.Shutdown, 10*time.Second)
}
