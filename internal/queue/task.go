package queue

import (
	"encoding/json"
	"fmt"
	"time"
)

// TaskType 任务类型
type TaskType string

const (
	// TaskTypePRReview PR 评审任务
	TaskTypePRReview TaskType = "pr_review"
)

// Task 任务数据结构
// 包含处理一个 PR 评审所需的所有信息
type Task struct {
	// 任务基本信息
	ID        string    // 任务唯一标识
	Type      TaskType  // 任务类型
	CreatedAt time.Time // 创建时间

	// PR 基本信息
	Owner    string // 仓库所有者（如 "gin-gonic"）
	Repo     string // 仓库名称（如 "gin"）
	PRNumber int    // PR 编号（如 123）
	SHA      string // PR 的最新 commit SHA

	// 事件信息
	Action   string // 触发事件类型（opened、synchronize、reopened）
	EventID  string // GitHub 事件 ID（用于去重）
}

// NewPRReviewTask 创建一个新的 PR 评审任务
func NewPRReviewTask(owner, repo string, prNumber int, sha, action, eventID string) *Task {
	return &Task{
		ID:        generateTaskID(owner, repo, prNumber, sha),
		Type:      TaskTypePRReview,
		CreatedAt: time.Now(),
		Owner:     owner,
		Repo:      repo,
		PRNumber:  prNumber,
		SHA:       sha,
		Action:    action,
		EventID:   eventID,
	}
}

// generateTaskID 生成任务唯一标识
// 格式: pr_review-{owner}-{repo}-{pr_number}-{sha[:8]}
func generateTaskID(owner, repo string, prNumber int, sha string) string {
	shortSHA := sha
	if len(sha) > 8 {
		shortSHA = sha[:8]
	}
	return fmt.Sprintf("pr_review-%s-%s-%d-%s", owner, repo, prNumber, shortSHA)
}

// ToJSON 将任务转换为 JSON 字符串（用于存储到 Redis）
func (t *Task) ToJSON() (string, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("序列化任务失败: %w", err)
	}
	return string(data), nil
}

// FromJSON 从 JSON 字符串解析任务
func FromJSON(data string) (*Task, error) {
	var task Task
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		return nil, fmt.Errorf("反序列化任务失败: %w", err)
	}
	return &task, nil
}

// String 返回任务的字符串表示（用于日志）
func (t *Task) String() string {
	shortSHA := t.SHA
	if len(shortSHA) > 8 {
		shortSHA = shortSHA[:8]
	}

	return fmt.Sprintf("Task[%s]: %s/%s#%d (action=%s, sha=%s)",
		t.ID, t.Owner, t.Repo, t.PRNumber, t.Action, shortSHA)
}
