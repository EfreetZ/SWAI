// Package main 是应用入口
// 职责：组装依赖并启动服务，不包含任何业务逻辑
package main

import (
	"flag"
	"log"

	"github.com/EfreetZ/SWAI/projects/stage0-engineering-template/internal/config"
	"github.com/EfreetZ/SWAI/projects/stage0-engineering-template/internal/logger"
	"github.com/EfreetZ/SWAI/projects/stage0-engineering-template/internal/server"
)

func main() {
	// 命令行参数：配置文件路径
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	// 1. 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. 初始化日志
	logger.New(cfg.Log.Level, cfg.Log.Format)

	// 3. 创建并启动服务
	srv := server.New(cfg)
	if err := srv.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
