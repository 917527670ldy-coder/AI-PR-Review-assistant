package server

import "github.com/gin-gonic/gin"

func NewRouter() *gin.Engine {
    r := gin.Default()
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    return r
}
