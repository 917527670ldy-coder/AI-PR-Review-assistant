package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Analyzer AI 代码分析器
type Analyzer struct {
	apiKey string
	model  string
}

// NewAnalyzer 创建 AI 分析器
// apiKey: 百炼 API Key
// model: 模型名称，如 "qwen-plus", "qwen-turbo", "qwen-max"
func NewAnalyzer(apiKey, model string) *Analyzer {
	return &Analyzer{
		apiKey: apiKey,
		model:  model,
	}
}

// ReviewResult 评审结果
type ReviewResult struct {
	Summary     string         // PR 变更总结
	Risks       []RiskItem     // 风险代码识别
	Suggestions []Suggestion   // Review 建议
	Score       int            // 代码质量评分 (0-100)
}

// RiskItem 风险项
type RiskItem struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// Suggestion 改进建议
type Suggestion struct {
	File        string `json:"file"`
	Description string `json:"description"`
	CodeSnippet string `json:"code_snippet"`
}

// Analyze 分析 PR 代码变更
func (a *Analyzer) Analyze(ctx context.Context, prInfo string, diffs []string) (*ReviewResult, error) {
	// 构建分析提示词
	prompt := buildAnalysisPrompt(prInfo, diffs)

	// 调用百炼 API
	content, err := callBailianAPI(ctx, a.apiKey, a.model, systemPrompt, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI 分析失败: %w", err)
	}

	// 解析响应
	result := parseAnalysisResult(content)

	return result, nil
}

// systemPrompt 系统提示词
const systemPrompt = `你是一位资深的技术专家，专门负责代码评审。你的任务是分析 Pull Request 的代码变更，识别潜在问题并提供改进建议。

你需要从以下几个维度进行分析：

1. **变更总结**：简要总结本次 PR 的主要变更内容和目的
2. **风险识别**：识别可能存在的问题，包括：
   - Bug 或逻辑错误
   - 安全漏洞
   - 性能问题
   - 代码质量问题
3. **改进建议**：提供具体的代码改进建议
4. **质量评分**：对本次变更的整体质量打分（0-100分）

输出格式要求（JSON）：
{
  "summary": "变更总结（1-2句话）",
  "risks": [
    {
      "file": "文件路径",
      "line": 行号（如果能确定），
      "description": "风险描述",
      "severity": "high/medium/low"
    }
  ],
  "suggestions": [
    {
      "file": "文件路径",
      "description": "建议描述",
      "code_snippet": "建议的代码片段（可选）"
    }
  ],
  "score": 评分数字
}

注意事项：
- 保持客观、专业的态度
- 识别真正有意义的风险，避免过度报告
- 建议要具体、可操作
- 对于不确定的问题，标注为 "low" 严重程度`

// buildAnalysisPrompt 构建分析提示词
func buildAnalysisPrompt(prInfo string, diffs []string) string {
	prompt := fmt.Sprintf("请分析以下 Pull Request 的代码变更：\n\n")
	prompt += fmt.Sprintf("PR 信息：\n%s\n\n", prInfo)
	prompt += "代码变更（Diff 格式）：\n"

	for i, diff := range diffs {
		if i >= 10 { // 限制分析的文件数量，避免超出 token 限制
			prompt += fmt.Sprintf("\n... 还有 %d 个文件的变更未展示\n", len(diffs)-10)
			break
		}
		prompt += fmt.Sprintf("\n---\n%s\n", diff)
	}

	return prompt
}

// callBailianAPI 调用百炼 API
func callBailianAPI(ctx context.Context, apiKey, model, systemPrompt, userPrompt string) (string, error) {
	// 构建请求体
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.3,
		"max_tokens":  4000,
	}

	// 序列化请求
	reqJson, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions", bytes.NewBuffer(reqJson))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API 返回错误 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var respData struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if err := json.Unmarshal(body, &respData); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(respData.Choices) == 0 {
		return "", fmt.Errorf("API 返回空响应")
	}

	return respData.Choices[0].Message.Content, nil
}

// parseAnalysisResult 解析 AI 响应（JSON 格式）
func parseAnalysisResult(content string) *ReviewResult {
	// 尝试从响应中提取 JSON
	// AI 可能返回带有其他文本的响应，需要找到 JSON 部分
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd < jsonStart {
		// 无法找到有效的 JSON，返回原始内容作为总结
		return &ReviewResult{
			Summary:     content,
			Risks:       []RiskItem{},
			Suggestions: []Suggestion{},
			Score:       80,
		}
	}

	jsonStr := content[jsonStart : jsonEnd+1]

	// 定义临时结构用于解析
	var rawResult struct {
		Summary     string `json:"summary"`
		Risks       []RiskItem `json:"risks"`
		Suggestions []Suggestion `json:"suggestions"`
		Score       int    `json:"score"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &rawResult); err != nil {
		// JSON 解析失败，返回原始内容作为总结
		return &ReviewResult{
			Summary:     fmt.Sprintf("解析失败: %v\n原始内容:\n%s", err, content),
			Risks:       []RiskItem{},
			Suggestions: []Suggestion{},
			Score:       80,
		}
	}

	return &ReviewResult{
		Summary:     rawResult.Summary,
		Risks:       rawResult.Risks,
		Suggestions: rawResult.Suggestions,
		Score:       rawResult.Score,
	}
}