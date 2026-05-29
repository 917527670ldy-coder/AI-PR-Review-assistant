package main

import (
	"fmt"
	"log"
	"os"

	"xengineer/internal/config"
	"xengineer/internal/github"
	"xengineer/internal/server"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 如果提供了命令行参数，则测试 GitHub API
	if len(os.Args) > 1 && os.Args[1] == "test-github" {
		testGitHubAPI(cfg)
		return
	}

	// 启动服务器
	log.Printf("正在启动服务器，端口: %s...", cfg.ServerPort)
	if err := server.Start(cfg); err != nil {
		log.Fatal(err)
	}
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
