package server

import (
    "github.com/gin-gonic/gin"
    "xengineer/internal/webhook"
)

func NewRouter() *gin.Engine {
    r := gin.Default()
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    r.POST("/webhook/github", webhook.HandleGitHubWebhook)
    return r
}
