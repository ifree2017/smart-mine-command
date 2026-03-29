package ai

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestNewAnalyzer(t *testing.T) {
	c := NewClient("https://api.openai.com", "token")
	a := NewAnalyzer(c)
	if a.client != c {
		t.Error("client not set")
	}
}

func TestNewAnalyzer_WithMock(t *testing.T) {
	mock := &mockLLMClient{resp: "{}", err: nil}
	a := NewAnalyzer(mock)
	if a.client != mock {
		t.Error("client not set")
	}
}

func TestAnalysisResult_Fields(t *testing.T) {
	r := AnalysisResult{
		Summary:        "test summary",
		RiskLevel:      3,
		Factors:        []string{"factor1", "factor2"},
		Recommendation: "建议测试",
	}
	if r.Summary != "test summary" {
		t.Errorf("Summary: got %s", r.Summary)
	}
	if r.RiskLevel != 3 {
		t.Errorf("RiskLevel: got %d", r.RiskLevel)
	}
	if len(r.Factors) != 2 {
		t.Errorf("Factors len: got %d", len(r.Factors))
	}
}

func TestAnalysisResult_DefaultRiskLevel(t *testing.T) {
	// 当 JSON 解析失败时，默认 riskLevel=3
	r := AnalysisResult{RiskLevel: 3}
	if r.RiskLevel != 3 {
		t.Errorf("default risk: got %d", r.RiskLevel)
	}
}

func TestAnalyzeAlert_Success(t *testing.T) {
	respJSON, _ := json.Marshal(AnalysisResult{
		Summary:        "高风险",
		RiskLevel:      5,
		Factors:        []string{"CH4浓度超标", "通风不良"},
		Recommendation: "立即撤离",
	})
	mock := &mockLLMClient{resp: string(respJSON), err: nil}
	a := NewAnalyzer(mock)

	alert := map[string]interface{}{
		"location": "001",
		"gasType":  "CH4",
		"value":    2.5,
		"level":    5,
	}

	result, err := a.AnalyzeAlert(context.Background(), alert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "高风险" {
		t.Errorf("Summary: got %s", result.Summary)
	}
	if result.RiskLevel != 5 {
		t.Errorf("RiskLevel: got %d", result.RiskLevel)
	}
	if len(result.Factors) != 2 {
		t.Errorf("Factors len: got %d", len(result.Factors))
	}
	if result.Recommendation != "立即撤离" {
		t.Errorf("Recommendation: got %s", result.Recommendation)
	}
}

func TestAnalyzeAlert_ClientError(t *testing.T) {
	mock := &mockLLMClient{resp: "", err: errors.New("network error")}
	a := NewAnalyzer(mock)

	alert := map[string]interface{}{
		"location": "001",
		"gasType":  "CH4",
		"value":    2.5,
		"level":    5,
	}

	_, err := a.AnalyzeAlert(context.Background(), alert)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err.Error() != "network error" {
		t.Errorf("error: got %v", err)
	}
}

func TestAnalyzeAlert_JSONFallback(t *testing.T) {
	// JSON解析失败时，RiskLevel默认为3
	mock := &mockLLMClient{resp: "not valid json", err: nil}
	a := NewAnalyzer(mock)

	alert := map[string]interface{}{
		"location": "001",
		"gasType":  "CO",
		"value":    0.5,
		"level":    3,
	}

	result, err := a.AnalyzeAlert(context.Background(), alert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RiskLevel != 3 {
		t.Errorf("RiskLevel: got %d, want 3 (default)", result.RiskLevel)
	}
	if result.Summary != "not valid json" {
		t.Errorf("Summary: got %s", result.Summary)
	}
}

func TestAnalyzeTrend_Success(t *testing.T) {
	respJSON, _ := json.Marshal(AnalysisResult{
		Summary:        "上升趋势",
		RiskLevel:      4,
		Factors:        []string{"多点报警", "浓度持续上升"},
		Recommendation: "加强监控",
	})
	mock := &mockLLMClient{resp: string(respJSON), err: nil}
	a := NewAnalyzer(mock)

	alerts := []map[string]interface{}{
		{"location": "001", "gasType": "CH4", "value": 1.0},
		{"location": "002", "gasType": "CO", "value": 0.3},
	}

	result, err := a.AnalyzeTrend(context.Background(), alerts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "上升趋势" {
		t.Errorf("Summary: got %s", result.Summary)
	}
	if result.RiskLevel != 4 {
		t.Errorf("RiskLevel: got %d", result.RiskLevel)
	}
}

func TestAnalyzeTrend_ClientError(t *testing.T) {
	mock := &mockLLMClient{resp: "", err: errors.New("timeout")}
	a := NewAnalyzer(mock)

	_, err := a.AnalyzeTrend(context.Background(), []map[string]interface{}{
		{"location": "001", "gasType": "CH4", "value": 1.0},
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestAnalyzeTrend_JSONFallback(t *testing.T) {
	mock := &mockLLMClient{resp: "raw text response", err: nil}
	a := NewAnalyzer(mock)

	result, err := a.AnalyzeTrend(context.Background(), []map[string]interface{}{
		{"location": "001", "gasType": "CH4", "value": 1.0},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RiskLevel != 3 {
		t.Errorf("RiskLevel: got %d, want 3 (default)", result.RiskLevel)
	}
}
