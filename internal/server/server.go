package server

import (
	"log"

	"xengineer/internal/config"
	"xengineer/internal/queue"
)

// Start 启动 HTTP 服务器
func Start(cfg *config.Config, q queue.Queue) error {
	r := NewRouter(cfg, q)
	log.Printf("🚀 服务器启动，监听端口: %s", cfg.ServerPort)
	return r.Run(":" + cfg.ServerPort)
}
