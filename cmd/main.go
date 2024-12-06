package main

import (
	"log"

	"mywallet/internal/config"
	"mywallet/internal/routes"
	"mywallet/pkg/logger"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	l := logger.NewLogger()
	defer l.Sync()

	// 初始化API服务器
	r := routes.InitRouter()
	if err := r.Run(cfg.ServerPort); err != nil {
		l.Fatal("服务器启动失败", err)
	}
}
