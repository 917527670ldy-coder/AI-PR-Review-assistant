package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xengineer/internal/config"
	"xengineer/internal/github"
	"xengineer/internal/queue"
	"xengineer/internal/worker"
)

func main() {
	fmt.Println("=== Worker 测试程序 ===")
	fmt.Println()

	// 1. 加载配置
	cfg := config.Load()
	if cfg.GitHubToken == "" {
		log.Fatal("❌ 请设置 GITHUB_TOKEN 环境变量")
	}
	fmt.Println("✅ 配置加载成功")

	// 2. 连接 Redis
	fmt.Println("连接 Redis...")
	q, err := queue.NewRedisQueue("localhost:6379", "pr_review_tasks")
	if err != nil {
		log.Fatalf("❌ 连接 Redis 失败: %v", err)
	}
	defer q.Close()
	fmt.Println("✅ Redis 连接成功")

	// 3. 创建 GitHub 客户端
	ghClient := github.NewClient(cfg.GitHubToken)
	fmt.Println("✅ GitHub 客户端创建成功")

	// 4. 创建 Worker
	w := worker.NewWorker(q, ghClient)
	fmt.Println("✅ Worker 创建成功")
	fmt.Println()

	// 5. 启动 Worker
	w.Start()

	// 6. 添加测试任务
	fmt.Println("添加测试任务...")
	testTask := queue.NewPRReviewTask(
		"gin-gonic", "gin", 1,
		"899a711ec71a7c0608f73a9ff898303db3997eb3",
		"opened", "test-event-001",
	)
	ctx := context.Background()
	if err := q.Enqueue(ctx, testTask); err != nil {
		log.Fatalf("❌ 入队失败: %v", err)
	}
	fmt.Printf("✅ 测试任务已入队: %s\n", testTask.String())
	fmt.Println()

	// 7. 等待 Worker 处理或用户中断
	fmt.Println("Worker 正在运行，等待任务处理...")
	fmt.Println("按 Ctrl+C 停止 Worker")
	fmt.Println()

	// 监听中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待一段时间后自动停止（或用户中断）
	timeout := time.After(30 * time.Second)
	select {
	case <-sigChan:
		fmt.Println("\n收到中断信号...")
	case <-timeout:
		fmt.Println("\n测试时间结束...")
	}

	// 8. 停止 Worker
	w.Stop()

	// 9. 输出统计信息
	stats := w.GetStats()
	fmt.Println()
	fmt.Println("=== Worker 统计信息 ===")
	fmt.Printf("已处理任务: %d\n", stats.Processed)
	fmt.Printf("失败任务: %d\n", stats.Failed)
	fmt.Printf("总耗时: %d 毫秒\n", stats.TotalTime)

	if stats.Processed > 0 {
		avgTime := stats.TotalTime / stats.Processed
		fmt.Printf("平均处理时间: %d 毫秒\n", avgTime)
	}

	fmt.Println()
	fmt.Println("✅ 测试完成")
}