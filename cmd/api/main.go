package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"xengineer/internal/ai"
	"xengineer/internal/config"
	"xengineer/internal/github"
	"xengineer/internal/queue"
	"xengineer/internal/server"
	"xengineer/internal/worker"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 如果提供了命令行参数，则测试 GitHub API
	if len(os.Args) > 1 && os.Args[1] == "test-github" {
		testGitHubAPI(cfg)
		return
	}

	// 启动完整系统
	startSystem(cfg)
}

// startSystem 启动完整系统
func startSystem(cfg *config.Config) {
	fmt.Println("=== AI PR Review Assistant ===")
	fmt.Println()

	// 1. 检查必要配置
	if cfg.GitHubToken == "" {
		log.Fatal("❌ 请设置 GITHUB_TOKEN 环境变量")
	}
	if cfg.ClaudeAPIKey == "" {
		log.Fatal("❌ 请设置 CLAUDE_API_KEY 环境变量")
	}
	if cfg.GitHubWebhookSecret == "" {
		log.Fatal("❌ 请设置 GITHUB_WEBHOOK_SECRET 环境变量")
	}
	fmt.Println("✅ 配置检查通过")

	// 2. 连接 Redis
	fmt.Println("连接 Redis...")
	q, err := queue.NewRedisQueue(cfg.RedisURL, "pr_review_tasks")
	if err != nil {
		log.Fatalf("❌ 连接 Redis 失败: %v", err)
	}
	defer q.Close()
	fmt.Println("✅ Redis 连接成功")

	// 3. 创建 GitHub 客户端
	ghClient := github.NewClient(cfg.GitHubToken)
	fmt.Println("✅ GitHub 客户端创建成功")

	// 4. 创建 AI 分析器
	analyzer := ai.NewAnalyzer(cfg.ClaudeAPIKey, cfg.AIModel)
	fmt.Println("✅ AI 分析器创建成功")

	// 5. 创建并启动 Worker
	w := worker.NewWorker(q, ghClient, analyzer)
	w.Start()
	fmt.Println("✅ Worker 已启动")
	fmt.Println()

	// 6. 启动 HTTP 服务器
	go func() {
		if err := server.Start(cfg, q); err != nil {
			log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 7. 等待中断信号
	fmt.Println("系统运行中...")
	fmt.Println("按 Ctrl+C 停止系统")
	fmt.Println()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 8. 停止 Worker
	fmt.Println("\n正在停止系统...")
	w.Stop()

	// 9. 输出统计信息
	stats := w.GetStats()
	fmt.Println()
	fmt.Println("=== Worker 统计信息 ===")
	fmt.Printf("已处理任务: %d\n", stats.Processed)
	fmt.Printf("失败任务: %d\n", stats.Failed)
	fmt.Printf("总耗时: %d 毫秒\n", stats.TotalTime)

	fmt.Println()
	fmt.Println("✅ 系统已停止")
}

// testGitHubAPI 测试 GitHub API 客户端
func testGitHubAPI(cfg *config.Config) {
	if cfg.GitHubToken == "" {
		log.Fatal("请设置 GITHUB_TOKEN 环境变量")
	}

	// 创建 GitHub 客户端
	client := github.NewClient(cfg.GitHubToken)

	// 测试获取 PR 信息
	// 这里使用一个公开的 PR 作为示例
	owner := "gin-gonic"
	repo := "gin"
	number := 1

	fmt.Printf("正在获取 PR: %s/%s#%d\n", owner, repo, number)

	pr, err := client.GetPR(owner, repo, number)
	if err != nil {
		log.Fatalf("获取 PR 失败: %v", err)
	}

	fmt.Printf("\n=== PR 信息 ===\n")
	fmt.Printf("标题: %s\n", pr.Title)
	fmt.Printf("作者: %s\n", pr.User)
	fmt.Printf("基础分支: %s\n", pr.BaseRef)
	fmt.Printf("目标分支: %s\n", pr.HeadRef)
	fmt.Printf("SHA: %s\n", pr.SHA)

	// 获取文件变更
	diffs, err := client.GetPRDiff(owner, repo, number)
	if err != nil {
		log.Fatalf("获取 PR Diff 失败: %v", err)
	}

	fmt.Printf("\n=== 文件变更 (%d 个文件) ===\n", len(diffs))
	for i, diff := range diffs {
		if i >= 5 {
			fmt.Printf("... 还有 %d 个文件\n", len(diffs)-5)
			break
		}
		fmt.Printf("- %s (%s): +%d -%d\n", diff.Filename, diff.Status, diff.Additions, diff.Deletions)
	}

	fmt.Println("\n✅ GitHub API 客户端测试成功！")
}
