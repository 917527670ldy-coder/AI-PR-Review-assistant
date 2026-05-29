package server

import (
	"github.com/gin-gonic/gin"
	"xengineer/internal/config"
	"xengineer/internal/webhook"
)

// NewRouter 创建并配置 Gin 路由
func NewRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Webhook 端点
	r.POST("/webhook/github", webhook.HandleGitHubWebhook)

	return r
}
