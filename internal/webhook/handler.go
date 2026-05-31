package webhook

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"xengineer/internal/queue"
)

// Handler Webhook 处理器
type Handler struct {
	queue  queue.Queue
	secret string
}

// NewHandler 创建 Webhook 处理器
func NewHandler(q queue.Queue, secret string) *Handler {
	return &Handler{
		queue:  q,
		secret: secret,
	}
}

// HandleGitHubWebhook 处理 GitHub Webhook 请求
func (h *Handler) HandleGitHubWebhook(c *gin.Context) {
	// 1. 获取事件类型和签名
	eventName := c.GetHeader("X-GitHub-Event")
	signature := c.GetHeader("X-Hub-Signature-256")
	deliveryID := c.GetHeader("X-GitHub-Delivery")

	// 2. 读取请求体
	body, err := c.GetRawData()
	if err != nil {
		log.Printf("⚠️  读取请求体失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// 3. 验证签名
	if !VerifySignature(h.secret, body, signature) {
		log.Printf("⚠️  Webhook 签名验证失败: delivery_id=%s", deliveryID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook signature"})
		return
	}

	// 4. 检查事件类型
	if eventName != "pull_request" {
		log.Printf("ℹ️  忽略非 PR 事件: event=%s, delivery_id=%s", eventName, deliveryID)
		c.JSON(http.StatusAccepted, gin.H{"status": "ignored event", "event": eventName})
		return
	}

	// 5. 解析 PR 事件
	var event PullRequestEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("⚠️  解析 PR 事件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pull_request payload"})
		return
	}

	// 6. 检查 Action 类型
	if event.Action != "opened" && event.Action != "synchronize" && event.Action != "reopened" {
		log.Printf("ℹ️  忽略 Action: repo=%s pr=%d action=%s",
			event.Repository.FullName, event.PullRequest.Number, event.Action)
		c.JSON(http.StatusAccepted, gin.H{"status": "ignored action", "action": event.Action})
		return
	}

	// 7. 解析仓库 Owner 和 Repo
	fullName := event.Repository.FullName
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		log.Printf("⚠️  无效的仓库全名: %s", fullName)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repository full_name"})
		return
	}
	owner := parts[0]
	repo := parts[1]

	// 8. 获取 PR 信息
	prNumber := event.PullRequest.Number
	if prNumber == 0 {
		prNumber = event.Number
	}
	sha := event.PullRequest.Head.SHA

	// 9. 创建评审任务
	task := queue.NewPRReviewTask(
		owner,
		repo,
		prNumber,
		sha,
		event.Action,
		deliveryID,
	)

	// 10. 将任务放入队列
	ctx := context.Background()
	if err := h.queue.Enqueue(ctx, task); err != nil {
		log.Printf("❌ 任务入队失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue task"})
		return
	}

	// 11. 返回成功响应
	log.Printf("✅ PR 评审任务已入队: repo=%s/%s pr=%d action=%s sha=%s task_id=%s",
		owner, repo, prNumber, event.Action, sha, task.ID)

	c.JSON(http.StatusAccepted, gin.H{
		"status":   "accepted",
		"task_id":  task.ID,
		"repo":     fullName,
		"pr_number": prNumber,
		"action":   event.Action,
	})
}

// HandleGitHubWebhook 兼容旧版本的函数（用于路由注册）
func HandleGitHubWebhook(c *gin.Context) {
	// 这个函数会被新的 Handler 替代
	// 这里保留是为了兼容性，实际使用时应该使用 Handler.HandleGitHubWebhook
	c.JSON(http.StatusOK, gin.H{"status": "handler not initialized"})
}