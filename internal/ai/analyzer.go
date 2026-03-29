package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

// Analyzer 安全分析器
type Analyzer struct {
	client *Client
}

// NewAnalyzer 创建分析器
func NewAnalyzer(client *Client) *Analyzer {
	return &Analyzer{client: client}
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	Summary       string   `json:"summary"`
	RiskLevel     int      `json:"risk_level"`  // 1-5
	Factors       []string `json:"factors"`
	Recommendation string  `json:"recommendation"`
}

// AnalyzeAlert 分析单条报警
func (a *Analyzer) AnalyzeAlert(ctx context.Context, alert map[string]interface{}) (*AnalysisResult, error) {
	prompt := fmt.Sprintf(`分析以下矿井报警，给出安全评估：

报警信息：
- 位置：%v
- 气体类型：%v
- 浓度值：%v
- 风险等级：%v

请用JSON格式回复，包含：
- summary: 简短总结
- risk_level: 1-5风险等级
- factors: 影响因素列表
- recommendation: 处理建议（50字内）
`,
		alert["location"], alert["gasType"], alert["value"], alert["level"])

	messages := []ChatMessage{
		{Role: "system", Content: "你是矿井安全专家。"},
		{Role: "user", Content: prompt},
	}

	resp, err := a.client.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	var result AnalysisResult
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		// 如果JSON解析失败，尝试提取
		result = AnalysisResult{
			Summary:       resp,
			RiskLevel:     3,
			Recommendation: "建议现场检查",
		}
	}

	return &result, nil
}

// AnalyzeTrend 分析趋势（多报警）
func (a *Analyzer) AnalyzeTrend(ctx context.Context, alerts []map[string]interface{}) (*AnalysisResult, error) {
	prompt := fmt.Sprintf(`分析以下矿井报警趋势，给出整体安全评估（共%v条报警）：

报警摘要：
`,
		len(alerts))

	for i, alert := range alerts {
		if i >= 10 { // 最多分析10条
			break
		}
		prompt += fmt.Sprintf("%v. %v - %v: %v\n",
			i+1, alert["location"], alert["gasType"], alert["value"])
	}

	prompt += `
请用JSON格式回复：
- summary: 趋势总结
- risk_level: 1-5整体风险等级
- factors: 主要影响因素
- recommendation: 综合建议`

	messages := []ChatMessage{
		{Role: "system", Content: "你是矿井安全专家。"},
		{Role: "user", Content: prompt},
	}

	resp, err := a.client.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	var result AnalysisResult
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		result = AnalysisResult{
			Summary:       resp,
			RiskLevel:     3,
			Recommendation: "持续监控",
		}
	}

	return &result, nil
}
