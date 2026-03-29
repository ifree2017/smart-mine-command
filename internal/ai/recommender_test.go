package ai

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"smart-mine-command/internal/dispatch"
	"smart-mine-command/internal/model"
)

func TestNewRecommender(t *testing.T) {
	c := NewClient("https://api.openai.com", "token")
	a := NewAnalyzer(c)
	r := NewRecommender(a)
	if r.analyzer != a {
		t.Error("analyzer not set")
	}
}

func TestBuildPlans(t *testing.T) {
	c := NewClient("https://api.openai.com", "token")
	a := NewAnalyzer(c)
	r := NewRecommender(a)

	alert := model.Alert{ID: "a1", Level: 4, GasType: "T", Location: "001"}
	cmds := []model.Command{
		{ID: "c1", Type: "broadcast", Content: "test"},
	}

	plans := r.buildPlans(alert, cmds)
	if len(plans) != 1 {
		t.Fatalf("plans: got %d, want 1", len(plans))
	}
	if plans[0].Name != "L4预案" {
		t.Errorf("Name: got %s", plans[0].Name)
	}
	if plans[0].Trigger != "level>=4" {
		t.Errorf("Trigger: got %s", plans[0].Trigger)
	}
}

func TestBuildPlans_EmptyCommands(t *testing.T) {
	c := NewClient("https://api.openai.com", "token")
	a := NewAnalyzer(c)
	r := NewRecommender(a)

	plans := r.buildPlans(model.Alert{}, nil)
	if plans != nil {
		t.Errorf("plans: got %v, want nil", plans)
	}
}

func TestEnrichWithAI(t *testing.T) {
	c := NewClient("https://api.openai.com", "token")
	a := NewAnalyzer(c)
	r := NewRecommender(a)

	plans := []dispatch.Plan{
		{Name: "L3预案"},
	}
	result := &AnalysisResult{Recommendation: "加强通风"}

	enriched := r.enrichWithAI(plans, result)
	if enriched[0].Name != "L3预案 | AI建议:加强通风" {
		t.Errorf("enriched name: got %s", enriched[0].Name)
	}
}

func TestEnrichWithAI_EmptyPlans(t *testing.T) {
	c := NewClient("https://api.openai.com", "token")
	a := NewAnalyzer(c)
	r := NewRecommender(a)

	result := &AnalysisResult{Recommendation: "test"}
	enriched := r.enrichWithAI([]dispatch.Plan{}, result)
	if len(enriched) != 0 {
		t.Errorf("len: got %d", len(enriched))
	}
}

// mockLLMClientForRecommender wraps mockLLMClient for Recommend tests
type mockLLMClientForRecommender struct {
	mockLLMClient
}

func TestRecommend_L4AlertWithAI(t *testing.T) {
	respJSON, _ := json.Marshal(AnalysisResult{
		Summary:        "高风险",
		RiskLevel:      5,
		Factors:        []string{"浓度超标"},
		Recommendation: "紧急撤离",
	})
	mock := &mockLLMClient{resp: string(respJSON), err: nil}
	a := NewAnalyzer(mock)
	r := NewRecommender(a)

	alert := model.Alert{
		ID:       "alert-1",
		Level:    4,
		GasType:  "CH4",
		Location: "A1",
		Value:    2.5,
	}

	plans, err := r.Recommend(context.Background(), alert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plans) == 0 {
		t.Fatal("plans should not be empty")
	}
	// L4 >= 3, so AI enrichment should be applied
	if plans[0].Name == "" {
		t.Error("plan name should not be empty")
	}
}

func TestRecommend_LowLevelNoAI(t *testing.T) {
	// Level 1-2 should not call AI, so no mock response needed
	c := NewClient("https://api.openai.com", "token")
	a := NewAnalyzer(c)
	r := NewRecommender(a)

	alert := model.Alert{
		ID:       "alert-2",
		Level:    2,
		GasType:  "CO",
		Location: "B2",
		Value:    0.1,
	}

	plans, err := r.Recommend(context.Background(), alert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Level < 3, should still have a plan from dispatch rules but no AI call
	if len(plans) == 0 {
		t.Fatal("plans should not be empty for L2 alert")
	}
}

func TestRecommend_AIErrorFallsBack(t *testing.T) {
	mock := &mockLLMClient{resp: "", err: errors.New("api error")}
	a := NewAnalyzer(mock)
	r := NewRecommender(a)

	alert := model.Alert{
		ID:       "alert-3",
		Level:    5,
		GasType:  "CH4",
		Location: "C3",
		Value:    3.0,
	}

	// Should not error, just fall back to rule-based plan
	plans, err := r.Recommend(context.Background(), alert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plans) == 0 {
		t.Fatal("plans should not be empty")
	}
}
