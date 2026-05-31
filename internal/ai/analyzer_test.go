package ai

import (
	"context"
	"testing"
	"time"
)

// TestAnalyzerCreation 测试 AI 分析器创建
func TestAnalyzerCreation(t *testing.T) {
	// 使用测试 API Key
	analyzer := NewAnalyzer("test-api-key", "qwen-plus")

	if analyzer.apiKey != "test-api-key" {
		t.Errorf("API Key不匹配")
	}
	if analyzer.model != "qwen-plus" {
		t.Errorf("模型名称不匹配")
	}

	t.Log("✅ AI 分析器创建成功")
}

// TestBuildAnalysisPrompt 测试分析提示词构建
func TestBuildAnalysisPrompt(t *testing.T) {
	prInfo := "标题: Test PR\n作者: test-user\n基础分支: main\n目标分支: feature"
	diffs := []string{
		"文件: test.go\n状态: modified\n变更: +10 -5\npatch content...",
		"文件: main.go\n状态: added\n变更: +20 -0\npatch content...",
	}

	prompt := buildAnalysisPrompt(prInfo, diffs)

	// 验证提示词包含必要信息
	if !contains(prompt, "Test PR") {
		t.Error("提示词应包含PR标题")
	}
	if !contains(prompt, "test-user") {
		t.Error("提示词应包含作者")
	}
	if !contains(prompt, "test.go") {
		t.Error("提示词应包含文件名")
	}

	t.Log("✅ 分析提示词构建正确")
}

// TestParseAnalysisResult 测试解析 AI 响应
func TestParseAnalysisResult(t *testing.T) {
	// 测试 JSON 格式响应
	jsonResponse := `{
		"summary": "本次 PR 添加了新功能",
		"risks": [
			{
				"file": "test.go",
				"line": 10,
				"description": "可能存在空指针问题",
				"severity": "medium"
			}
		],
		"suggestions": [
			{
				"file": "test.go",
				"description": "建议添加错误处理",
				"code_snippet": "if err != nil { return err }"
			}
		],
		"score": 85
	}`

	result := parseAnalysisResult(jsonResponse)

	if result.Summary != "本次 PR 添加了新功能" {
		t.Errorf("Summary不匹配: %s", result.Summary)
	}
	if len(result.Risks) != 1 {
		t.Errorf("应识别1个风险，实际: %d", len(result.Risks))
	}
	if result.Risks[0].Severity != "medium" {
		t.Errorf("风险严重程度应为medium，实际: %s", result.Risks[0].Severity)
	}
	if len(result.Suggestions) != 1 {
		t.Errorf("应提供1个建议，实际: %d", len(result.Suggestions))
	}
	if result.Score != 85 {
		t.Errorf("评分应为85，实际: %d", result.Score)
	}

	t.Log("✅ AI 响应解析正确")
}

// TestParseNonJSONResponse 测试解析非 JSON 响应
func TestParseNonJSONResponse(t *testing.T) {
	nonJSONResponse := "这是一个普通的文本响应，不包含JSON格式"

	result := parseAnalysisResult(nonJSONResponse)

	// 非JSON响应应返回原始内容作为总结
	if result.Summary != nonJSONResponse {
		t.Errorf("非JSON响应应作为Summary: %s", result.Summary)
	}
	if len(result.Risks) != 0 {
		t.Errorf("应无风险，实际: %d", len(result.Risks))
	}
	if len(result.Suggestions) != 0 {
		t.Errorf("应无建议，实际: %d", len(result.Suggestions))
	}
	if result.Score != 80 {
		t.Errorf("默认评分应为80，实际: %d", result.Score)
	}

	t.Log("✅ 非JSON响应处理正确")
}

// TestAnalyzeWithRealAPI 测试真实 API 调用（需要 API Key）
// 这个测试需要设置环境变量 CLAUDE_API_KEY
func TestAnalyzeWithRealAPI(t *testing.T) {
	// 检查是否设置了 API Key
	apiKey := getEnvOrDefault("CLAUDE_API_KEY", "")
	if apiKey == "" {
		t.Skip("未设置 CLAUDE_API_KEY，跳过真实API测试")
	}

	analyzer := NewAnalyzer(apiKey, "qwen-plus")

	// 构建测试数据
	prInfo := "标题: 修复中间件索引越界问题\n作者: test-user\n基础分支: master\n目标分支: feature"
	diffs := []string{
		"文件: middleware.go\n状态: modified\n变更: +2 -2\n@@ -45,7 +45,7 @@ func(c *Context) {\n-    arr[len(arr)]\n+    arr[len(arr)-1]\n}",
	}

	// 调用真实API（设置较短超时）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := analyzer.Analyze(ctx, prInfo, diffs)
	if err != nil {
		t.Fatalf("AI 分析失败: %v", err)
	}

	// 验证结果
	if result.Summary == "" {
		t.Error("Summary不应为空")
	}
	if result.Score < 0 || result.Score > 100 {
		t.Errorf("评分应在0-100之间，实际: %d", result.Score)
	}

	t.Logf("✅ AI 分析成功")
	t.Logf("   变更总结: %s", result.Summary)
	t.Logf("   风险数量: %d", len(result.Risks))
	t.Logf("   建议数量: %d", len(result.Suggestions))
	t.Logf("   质量评分: %d/100", result.Score)
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := getEnv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnv 获取环境变量（这里简化实现）
func getEnv(key string) string {
	// 在实际测试中会从 os.Getenv 获取
	return ""
}