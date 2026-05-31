package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"xengineer/internal/ai"
	"xengineer/internal/github"
	"xengineer/internal/queue"
)

// Worker 后台任务处理器
// 从队列中取出任务并执行 PR 评审流程
type Worker struct {
	// 依赖的服务
	queue  queue.Queue
	github *github.Client
	ai     *ai.Analyzer // AI 分析器

	// 配置参数
	pollTimeout time.Duration // 每次等待任务的超时时间
	maxRetries  int           // 任务失败时的最大重试次数

	// 控制
	ctx    context.Context
	cancel context.CancelFunc

	// 状态
	running bool
	mu      sync.Mutex

	// 统计
	stats *Stats
}

// Stats Worker 统计信息
type Stats struct {
	Processed   int64 // 已处理任务数
	Failed      int64 // 失败任务数
	TotalTime   int64 // 总处理时间（毫秒）
	mu          sync.Mutex
}

// NewWorker 创建新的 Worker
func NewWorker(q queue.Queue, gh *github.Client, analyzer *ai.Analyzer) *Worker {
	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		queue:       q,
		github:      gh,
		ai:          analyzer,
		pollTimeout: 10 * time.Second, // 默认等待 10 秒
		maxRetries:  3,                // 默认重试 3 次
		ctx:         ctx,
		cancel:      cancel,
		running:     false,
		stats:       &Stats{},
	}
}

// Start 启动 Worker
// Worker 会开始从队列中取任务并处理
func (w *Worker) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return // 已经在运行
	}
	w.running = true
	w.mu.Unlock()

	log.Println("🚀 Worker 已启动，开始处理任务...")
	go w.run()
}

// Stop 停止 Worker
func (w *Worker) Stop() {
	w.mu.Lock()
	w.running = false
	w.mu.Unlock()

	w.cancel()
	log.Println("🛑 Worker 已停止")
}

// run Worker 主循环
func (w *Worker) run() {
	for {
		// 检查是否应该停止
		w.mu.Lock()
		shouldStop := !w.running
		w.mu.Unlock()

		if shouldStop {
			log.Println("Worker 主循环退出")
			return
		}

		// 从队列取任务
		task, err := w.queue.Dequeue(w.ctx, w.pollTimeout)
		if err != nil {
			if err == context.Canceled {
				return // Worker 被停止
			}
			log.Printf("⚠️  从队列取任务失败: %v", err)
			continue
		}

		if task == nil {
			// 队列为空，继续等待
			continue
		}

		// 处理任务
		w.processTask(task)
	}
}

// processTask 处理单个任务
func (w *Worker) processTask(task *queue.Task) {
	startTime := time.Now()
	log.Printf("📋 开始处理任务: %s", task.String())

	var err error

	// 重试机制
	for attempt := 1; attempt <= w.maxRetries; attempt++ {
		err = w.doProcess(task)
		if err == nil {
			break // 成功
		}

		log.Printf("⚠️  任务处理失败 (尝试 %d/%d): %v", attempt, w.maxRetries, err)

		if attempt < w.maxRetries {
			// 等待后重试
			waitTime := time.Duration(attempt) * time.Second
			log.Printf("⏳ %v 后重试...", waitTime)
			time.Sleep(waitTime)
		}
	}

	// 更新统计
	elapsed := time.Since(startTime)
	w.updateStats(err == nil, elapsed)

	if err != nil {
		log.Printf("❌ 任务最终失败: %s, 错误: %v", task.String(), err)
	} else {
		log.Printf("✅ 任务处理成功: %s, 耗时: %v", task.String(), elapsed)
	}
}

// doProcess 执行实际的任务处理逻辑
func (w *Worker) doProcess(task *queue.Task) error {
	// 步骤 1: 获取 PR 基本信息
	log.Printf("  → 步骤 1: 获取 PR 信息 (%s/%s#%d)", task.Owner, task.Repo, task.PRNumber)
	prInfo, err := w.github.GetPR(task.Owner, task.Repo, task.PRNumber)
	if err != nil {
		return fmt.Errorf("获取 PR 信息失败: %w", err)
	}
	log.Printf("     ✓ PR 标题: %s, 作者: %s", prInfo.Title, prInfo.User)

	// 步骤 2: 获取 PR 代码变更
	log.Printf("  → 步骤 2: 获取代码变更")
	diffs, err := w.github.GetPRDiff(task.Owner, task.Repo, task.PRNumber)
	if err != nil {
		return fmt.Errorf("获取代码变更失败: %w", err)
	}
	log.Printf("     ✓ 共 %d 个文件变更", len(diffs))

	// 步骤 3: 调用 AI 分析
	log.Printf("  → 步骤 3: AI 分析代码变更")

	// 构建 PR 信息字符串
	prInfoStr := fmt.Sprintf("标题: %s\n作者: %s\n基础分支: %s\n目标分支: %s",
		prInfo.Title, prInfo.User, prInfo.BaseRef, prInfo.HeadRef)

	// 构建 Diff 字符串列表
	var diffStrings []string
	for _, diff := range diffs {
		diffStr := fmt.Sprintf("文件: %s\n状态: %s\n变更: +%d -%d\n%s",
			diff.Filename, diff.Status, diff.Additions, diff.Deletions, diff.Patch)
		diffStrings = append(diffStrings, diffStr)
	}

	// 调用 AI 分析器
	reviewResult, err := w.ai.Analyze(w.ctx, prInfoStr, diffStrings)
	if err != nil {
		return fmt.Errorf("AI 分析失败: %w", err)
	}
	log.Printf("     ✓ AI 分析完成，质量评分: %d/100", reviewResult.Score)

	// 步骤 4: 发布评审结果
	log.Printf("  → 步骤 4: 发布评审评论")
	comment := buildReviewComment(prInfo, diffs, reviewResult)

	err = w.github.CreatePRComment(task.Owner, task.Repo, task.PRNumber, comment)
	if err != nil {
		return fmt.Errorf("发布评论失败: %w", err)
	}
	log.Printf("     ✓ 评审评论已发布")

	return nil
}

// buildReviewComment 构建评审评论内容
func buildReviewComment(prInfo *github.PRInfo, diffs []github.FileDiff, result *ai.ReviewResult) string {
	comment := fmt.Sprintf("## 🤖 AI 代码评审报告\n\n")
	comment += fmt.Sprintf("**PR #%d: %s**\n", prInfo.Number, prInfo.Title)
	comment += fmt.Sprintf("**作者:** %s\n\n", prInfo.User)

	// 变更总结
	comment += fmt.Sprintf("### 📌 变更总结\n\n%s\n\n", result.Summary)

	// 文件变更统计
	comment += fmt.Sprintf("### 📊 变更统计\n\n")
	comment += fmt.Sprintf("- 文件变更数: %d\n", len(diffs))
	totalAdditions := 0
	totalDeletions := 0
	for _, diff := range diffs {
		totalAdditions += diff.Additions
		totalDeletions += diff.Deletions
	}
	comment += fmt.Sprintf("- 新增行数: %d\n", totalAdditions)
	comment += fmt.Sprintf("- 删除行数: %d\n\n", totalDeletions)

	// 风险识别
	if len(result.Risks) > 0 {
		comment += fmt.Sprintf("### ⚠️ 风险识别\n\n")
		for i, risk := range result.Risks {
			severityEmoji := getSeverityEmoji(risk.Severity)
			comment += fmt.Sprintf("%d. %s **%s** (严重程度: %s)\n", i+1, severityEmoji, risk.Description, risk.Severity)
			if risk.File != "" {
				comment += fmt.Sprintf("   - 文件: `%s`\n", risk.File)
			}
			if risk.Line > 0 {
				comment += fmt.Sprintf("   - 行号: %d\n", risk.Line)
			}
			comment += "\n"
		}
	} else {
		comment += fmt.Sprintf("### ✅ 风险识别\n\n未发现明显风险问题。\n\n")
	}

	// 改进建议
	if len(result.Suggestions) > 0 {
		comment += fmt.Sprintf("### 💡 改进建议\n\n")
		for i, suggestion := range result.Suggestions {
			comment += fmt.Sprintf("%d. %s\n", i+1, suggestion.Description)
			if suggestion.File != "" {
				comment += fmt.Sprintf("   - 文件: `%s`\n", suggestion.File)
			}
			if suggestion.CodeSnippet != "" {
				comment += fmt.Sprintf("   ```\n   %s\n   ```\n", suggestion.CodeSnippet)
			}
			comment += "\n"
		}
	} else {
		comment += fmt.Sprintf("### 💡 改进建议\n\n暂无特别建议。\n\n")
	}

	// 质量评分
	scoreEmoji := getScoreEmoji(result.Score)
	comment += fmt.Sprintf("### 🎯 代码质量评分\n\n")
	comment += fmt.Sprintf("%s **%d/100**\n\n", scoreEmoji, result.Score)

	// 评审时间
	comment += fmt.Sprintf("---\n\n")
	comment += fmt.Sprintf("_由 AI PR Review Bot 自动生成_\n")

	return comment
}

// getSeverityEmoji 根据严重程度返回对应的 Emoji
func getSeverityEmoji(severity string) string {
	switch severity {
	case "high":
		return "🔴"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	default:
		return "⚪"
	}
}

// getScoreEmoji 根据评分返回对应的 Emoji
func getScoreEmoji(score int) string {
	if score >= 90 {
		return "🌟"
	} else if score >= 80 {
		return "✨"
	} else if score >= 70 {
		return "👍"
	} else if score >= 60 {
		return "😐"
	} else {
		return "⚠️"
	}
}

// updateStats 更新统计信息
func (w *Worker) updateStats(success bool, elapsed time.Duration) {
	w.stats.mu.Lock()
	defer w.stats.mu.Unlock()

	if success {
		w.stats.Processed++
	} else {
		w.stats.Failed++
	}
	w.stats.TotalTime += elapsed.Milliseconds()
}

// GetStats 获取统计信息
func (w *Worker) GetStats() Stats {
	w.stats.mu.Lock()
	defer w.stats.mu.Unlock()

	return Stats{
		Processed: w.stats.Processed,
		Failed:    w.stats.Failed,
		TotalTime: w.stats.TotalTime,
	}
}

// IsRunning 检查 Worker 是否在运行
func (w *Worker) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}