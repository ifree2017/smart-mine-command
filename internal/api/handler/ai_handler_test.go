package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"smart-mine-command/internal/ai"
	"smart-mine-command/internal/dispatch"
)

// mockLLMClient implements ai.LLMClient for testing.
type mockLLMClient struct {
	resp string
	err  error
}

func (m *mockLLMClient) Chat(ctx context.Context, messages []ai.ChatMessage) (string, error) {
	return m.resp, m.err
}

// makeTestAnalyzer builds a *ai.Analyzer wired to mockLLMClient.
func makeTestAnalyzer(resp string, err error) *ai.Analyzer {
	return ai.NewAnalyzer(&mockLLMClient{resp: resp, err: err})
}

// makeTestRecommender builds a *ai.Recommender with a real recommender
// driven by the given mock LLMClient (via analyzer).
func makeTestRecommender(llmResp string, llmErr error) *ai.Recommender {
	analyzer := ai.NewAnalyzer(&mockLLMClient{resp: llmResp, err: llmErr})
	return ai.NewRecommender(analyzer)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestAnalyzeHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyzer := makeTestAnalyzer(`{"summary":"test分析","risk_level":3,"factors":["因素1"],"recommendation":"建议测试"}`, nil)
	recommender := makeTestRecommender("", nil)
	h := NewAIHandler(analyzer, recommender)

	reqBody := `{"location":"001","gasType":"T","value":0.5,"level":3}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/ai/analyze", bytes.NewReader([]byte(reqBody)))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Analyze(c)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}

	var resp ai.AnalysisResult
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Summary != "test分析" {
		t.Errorf("summary: got %q, want %q", resp.Summary, "test分析")
	}
	if resp.RiskLevel != 3 {
		t.Errorf("risk_level: got %d, want 3", resp.RiskLevel)
	}
}

func TestAnalyzeHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyzer := makeTestAnalyzer("", nil)
	recommender := makeTestRecommender("", nil)
	h := NewAIHandler(analyzer, recommender)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/ai/analyze", bytes.NewReader([]byte("invalid")))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Analyze(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestAnalyzeHandler_AnalyzerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	analyzer := makeTestAnalyzer("", errors.New("llm error"))
	recommender := makeTestRecommender("", nil)
	h := NewAIHandler(analyzer, recommender)

	reqBody := `{"location":"001","gasType":"T","value":0.5,"level":3}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/ai/analyze", bytes.NewReader([]byte(reqBody)))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Analyze(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

// TestRecommendHandler_OK_LowLevel tests Recommend with level<3 where AI is NOT called.
// dispatch.Evaluate with DefaultRules for level=2 matches {LevelMin:1} rule only.
func TestRecommendHandler_OK_LowLevel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recommender := makeTestRecommender("", nil) // LLM won't be called
	h := NewAIHandler(makeTestAnalyzer("", nil), recommender)

	// level=2: only {LevelMin:1} rule matches → ws notification
	reqBody := `{"location":"001","gasType":"T","value":0.5,"level":2}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/ai/recommend", bytes.NewReader([]byte(reqBody)))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Recommend(c)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}

	var resp struct {
		Plans []dispatch.Plan `json:"plans"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Plans) != 1 {
		t.Fatalf("plans count: got %d, want 1", len(resp.Plans))
	}
	if resp.Plans[0].ID != "plan-2" {
		t.Errorf("plan ID: got %q, want %q", resp.Plans[0].ID, "plan-2")
	}
	if resp.Plans[0].Name != "L2预案" {
		t.Errorf("plan Name: got %q, want %q", resp.Plans[0].Name, "L2预案")
	}
}

// TestRecommendHandler_OK_HighLevel tests Recommend with level>=3 where AI IS called.
// dispatch.Evaluate matches {LevelMin:3} and {LevelMin:1} rules.
// AI recommendation is appended via enrichWithAI.
func TestRecommendHandler_OK_HighLevel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	llmResp := `{"summary":"test","risk_level":3,"factors":["f1"],"recommendation":"现场检查"}`
	recommender := makeTestRecommender(llmResp, nil)
	h := NewAIHandler(makeTestAnalyzer(llmResp, nil), recommender)

	reqBody := `{"location":"A1","gasType":"CH4","value":1.5,"level":3}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/ai/recommend", bytes.NewReader([]byte(reqBody)))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Recommend(c)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}

	var resp struct {
		Plans []dispatch.Plan `json:"plans"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Plans) != 1 {
		t.Fatalf("plans count: got %d, want 1", len(resp.Plans))
	}
	// Name enriched by AI
	if resp.Plans[0].Name != "L3预案 | AI建议:现场检查" {
		t.Errorf("plan Name: got %q, want %q", resp.Plans[0].Name, "L3预案 | AI建议:现场检查")
	}
}

func TestRecommendHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAIHandler(makeTestAnalyzer("", nil), makeTestRecommender("", nil))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/ai/recommend", bytes.NewReader([]byte("invalid")))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Recommend(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

// TestRecommendHandler_AIError tests that Recommend still returns 200 with valid
// plans when the AI analyzer fails (error is logged but non-fatal).
func TestRecommendHandler_AIError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recommender := makeTestRecommender("", errors.New("llm error"))
	h := NewAIHandler(makeTestAnalyzer("", nil), recommender)

	reqBody := `{"location":"A1","gasType":"CH4","value":1.5,"level":3}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/ai/recommend", bytes.NewReader([]byte(reqBody)))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Recommend(c)

	// AI error is non-fatal; should still return 200 with base plan
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}

	var resp struct {
		Plans []dispatch.Plan `json:"plans"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Plans) != 1 {
		t.Fatalf("plans count: got %d, want 1", len(resp.Plans))
	}
	// No AI enrichment on error
	if resp.Plans[0].Name != "L3预案" {
		t.Errorf("plan Name: got %q, want %q (no AI enrichment)", resp.Plans[0].Name, "L3预案")
	}
}

// TestRecommendHandler_HighLevel_NoPlans tests that Recommend returns 200
// with empty plans when dispatch.Evaluate produces no commands.
func TestRecommendHandler_NoPlans(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAIHandler(makeTestAnalyzer("", nil), makeTestRecommender("", nil))

	// dispatch.Evaluate with DefaultRules for level=1 matches {LevelMin:1} → ws notification
	// For truly no matches, use an alert level of 0 (no rule matches level 0)
	reqBody := `{"location":"A1","gasType":"CO","value":0.01,"level":0}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/ai/recommend", bytes.NewReader([]byte(reqBody)))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Recommend(c)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}

	var resp struct {
		Plans []dispatch.Plan `json:"plans"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Plans != nil {
		t.Errorf("plans: got %v, want nil", resp.Plans)
	}
}
