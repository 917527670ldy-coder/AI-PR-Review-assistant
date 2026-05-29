package server

import (
	"xengineer/internal/config"
)

// Start 启动 HTTP 服务器
func Start(cfg *config.Config) error {
	r := NewRouter(cfg)
	return r.Run(":" + cfg.ServerPort)
}
