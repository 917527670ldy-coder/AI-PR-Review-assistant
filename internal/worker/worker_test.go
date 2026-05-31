package worker

import (
	"strings"
	"testing"

	"xengineer/internal/ai"
	"xengineer/internal/github"
)

// TestWorkerCreation 测试 Worker 创建
func TestWorkerCreation(t *testing.T) {
	// 注意：这个测试需要实际的 Queue 和 GitHub Client
	// 在实际环境中，应该使用 Mock 对象
	t.Log("✅ Worker 创建测试（需要实际依赖）")
}

// TestWorkerStats 测试统计信息
func TestWorkerStats(t *testing.T) {
	stats := &Stats{
		Processed: 10,
		Failed:    2,
		TotalTime: 5000,
	}

	// 验证统计信息
	if stats.Processed != 10 {
		t.Errorf("Processed 应为 10，实际: %d", stats.Processed)
	}
	if stats.Failed != 2 {
		t.Errorf("Failed 应为 2，实际: %d", stats.Failed)
	}
	if stats.TotalTime != 5000 {
		t.Errorf("TotalTime 应为 5000，实际: %d", stats.TotalTime)
	}

	t.Logf("✅ 统计信息正确: 处理 %d，失败 %d，总耗时 %dms", stats.Processed, stats.Failed, stats.TotalTime)
}

// TestSeverityEmoji 测试严重程度 Emoji
func TestSeverityEmoji(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{"high", "🔴"},
		{"medium", "🟡"},
		{"low", "🟢"},
		{"unknown", "⚪"},
	}

	for _, test := range tests {
		emoji := getSeverityEmoji(test.severity)
		if emoji != test.expected {
			t.Errorf("严重程度 %s 的 Emoji 应为 %s，实际: %s", test.severity, test.expected, emoji)
		}
	}

	t.Log("✅ 严重程度 Emoji 正确")
}

// TestScoreEmoji 测试评分 Emoji
func TestScoreEmoji(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{95, "🌟"},
		{85, "✨"},
		{75, "👍"},
		{65, "😐"},
		{55, "⚠️"},
	}

	for _, test := range tests {
		emoji := getScoreEmoji(test.score)
		if emoji != test.expected {
			t.Errorf("评分 %d 的 Emoji 应为 %s，实际: %s", test.score, test.expected, emoji)
		}
	}

	t.Log("✅ 评分 Emoji 正确")
}

// TestBuildReviewComment 测试评审评论构建
func TestBuildReviewComment(t *testing.T) {
	// 创建测试数据
	prInfo := &github.PRInfo{
		Number:  1,
		Title:   "Test PR",
		User:    "test-user",
		BaseRef: "main",
		HeadRef: "feature",
	}

	diffs := []github.FileDiff{
		{
			Filename:  "test.go",
			Status:    "modified",
			Additions: 10,
			Deletions: 5,
		},
	}

	result := &ai.ReviewResult{
		Summary: "本次 PR 添加了新功能",
		Risks: []ai.RiskItem{
			{
				File:        "test.go",
				Line:        10,
				Description: "可能存在空指针问题",
				Severity:    "medium",
			},
		},
		Suggestions: []ai.Suggestion{
			{
				File:        "test.go",
				Description: "建议添加错误处理",
				CodeSnippet: "if err != nil { return err }",
			},
		},
		Score: 85,
	}

	comment := buildReviewComment(prInfo, diffs, result)

	// 验证评论包含必要信息
	if !strings.Contains(comment, "Test PR") {
		t.Error("评论应包含 PR 标题")
	}
	if !strings.Contains(comment, "test-user") {
		t.Error("评论应包含作者")
	}
	if !strings.Contains(comment, "新功能") {
		t.Error("评论应包含变更总结")
	}
	if !strings.Contains(comment, "空指针问题") {
		t.Error("评论应包含风险描述")
	}
	if !strings.Contains(comment, "错误处理") {
		t.Error("评论应包含改进建议")
	}
	if !strings.Contains(comment, "85") {
		t.Error("评论应包含质量评分")
	}

	t.Log("✅ 评审评论构建正确")
}