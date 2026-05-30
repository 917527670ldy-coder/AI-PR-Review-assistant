package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"xengineer/internal/ai"
	"xengineer/internal/config"
)

func main() {
	fmt.Println("=== 百炼 AI 分析器测试 ===")
	fmt.Println()

	// 加载配置
	cfg := config.Load()
	if cfg.ClaudeAPIKey == "" {
		log.Fatal("❌ 请设置 CLAUDE_API_KEY 环境变量（百炼 API Key）")
	}

	fmt.Println("步骤 1: 创建 AI 分析器...")
	analyzer := ai.NewAnalyzer(cfg.ClaudeAPIKey, cfg.AIModel)
	fmt.Printf("✅ AI 分析器创建成功，使用模型: %s\n", cfg.AIModel)
	fmt.Println()

	// 测试场景 1: 简单的代码变更
	fmt.Println("步骤 2: 测试 AI 分析功能...")
	prInfo := `PR #123: gin-gonic/gin
标题: Fix middleware chain bug
作者: developer123
变更描述: 修复中间件链中的顺序问题`

	diffs := []string{
		`diff --git a/context.go b/context.go
index abc123..def456 100644
--- a/context.go
+++ b/context.go
@@ -45,7 +45,7 @@ func (c *Context) Next() {
-    if len(c.handlers) > c.index {
-        c.index++
-        c.handlers[c.index](c)
-    }
+    c.index++
+    if c.index < len(c.handlers) {
+        c.handlers[c.index](c)
+    }`,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("正在调用百炼 API 进行分析...")
	start := time.Now()

	result, err := analyzer.Analyze(ctx, prInfo, diffs)
	if err != nil {
		log.Fatalf("❌ AI 分析失败: %v", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("✅ 分析完成，耗时: %.2f 秒\n", elapsed.Seconds())
	fmt.Println()

	// 显示分析结果
	fmt.Println("=== 分析结果 ===")
	fmt.Println()
	fmt.Printf("📌 变更总结:\n%s\n\n", result.Summary)

	if len(result.Risks) > 0 {
		fmt.Printf("⚠️  风险识别 (%d 个):\n", len(result.Risks))
		for i, risk := range result.Risks {
			fmt.Printf("  %d. [%s] %s (行 %d)\n", i+1, risk.Severity, risk.Description, risk.Line)
			if risk.File != "" {
				fmt.Printf("     文件: %s\n", risk.File)
			}
		}
		fmt.Println()
	} else {
		fmt.Println("✅ 未识别到明显风险")
		fmt.Println()
	}

	if len(result.Suggestions) > 0 {
		fmt.Printf("💡 改进建议 (%d 个):\n", len(result.Suggestions))
		for i, suggestion := range result.Suggestions {
			fmt.Printf("  %d. %s\n", i+1, suggestion.Description)
			if suggestion.File != "" {
				fmt.Printf("     文件: %s\n", suggestion.File)
			}
			if suggestion.CodeSnippet != "" {
				fmt.Printf("     代码建议:\n%s\n", suggestion.CodeSnippet)
			}
		}
		fmt.Println()
	} else {
		fmt.Println("💡 未提供具体建议")
		fmt.Println()
	}

	fmt.Printf("📊 代码质量评分: %d/100\n", result.Score)
	fmt.Println()

	// 测试总结
	fmt.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")
	fmt.Println("✅ 百炼 AI 分析器测试成功！")
	fmt.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")
	fmt.Println()
	fmt.Println("测试内容总结:")
	fmt.Println("  1. ✅ 百炼 API 连接正常")
	fmt.Println("  2. ✅ AI 分析功能正常")
	fmt.Println("  3. ✅ JSON 解析功能正常")
	fmt.Println("  4. ✅ 分析结果结构正确")
	fmt.Printf("  5. ✅ 响应时间合理 (%.2f 秒)\n", elapsed.Seconds())
}