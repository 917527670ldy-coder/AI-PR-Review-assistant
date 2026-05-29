package webhook

import (
    "encoding/json"
    "log"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
)

func HandleGitHubWebhook(c *gin.Context) {
    eventName := c.GetHeader("X-GitHub-Event")
    signature := c.GetHeader("X-Hub-Signature-256")

    body, err := c.GetRawData()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
        return
    }

    secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
    if !VerifySignature(secret, body, signature) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook signature"})
        return
    }

    if eventName != "pull_request" {
        c.JSON(http.StatusAccepted, gin.H{"status": "ignored event", "event": eventName})
        return
    }

    var event PullRequestEvent
    if err := json.Unmarshal(body, &event); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pull_request payload"})
        return
    }

    if event.Action != "opened" && event.Action != "synchronize" && event.Action != "reopened" {
        c.JSON(http.StatusAccepted, gin.H{"status": "ignored action", "action": event.Action})
        return
    }

    prNumber := event.Number
    if prNumber == 0 {
        prNumber = event.PullRequest.Number
    }

    log.Printf("pull_request received: repo=%s pr_number=%d action=%s head_sha=%s",
        event.Repository.FullName,
        prNumber,
        event.Action,
        event.PullRequest.Head.SHA,
    )

    c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}
