package main

import (
	"log"

	"xengineer/internal/config"
	"xengineer/internal/server"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 启动服务器
	log.Printf("正在启动服务器，端口: %s...", cfg.ServerPort)
	if err := server.Start(cfg); err != nil {
		log.Fatal(err)
	}
}
