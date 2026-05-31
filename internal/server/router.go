package server

import (
	"github.com/gin-gonic/gin"
	"xengineer/internal/config"
	"xengineer/internal/queue"
	"xengineer/internal/webhook"
)

// NewRouter 创建并配置 Gin 路由
func NewRouter(cfg *config.Config, q queue.Queue) *gin.Engine {
	r := gin.Default()

	// 创建 Webhook Handler
	webhookHandler := webhook.NewHandler(q, cfg.GitHubWebhookSecret)

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Webhook 端点
	r.POST("/webhook/github", webhookHandler.HandleGitHubWebhook)

	return r
}
