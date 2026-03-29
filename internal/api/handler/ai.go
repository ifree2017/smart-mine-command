package handler

import (
	"github.com/gin-gonic/gin"

	"smart-mine-command/internal/ai"
	"smart-mine-command/internal/model"
)

type AIHandler struct {
	analyzer    *ai.Analyzer
	recommender *ai.Recommender
}

func NewAIHandler(analyzer *ai.Analyzer, recommender *ai.Recommender) *AIHandler {
	return &AIHandler{
		analyzer:    analyzer,
		recommender: recommender,
	}
}

func (h *AIHandler) Analyze(c *gin.Context) {
	var req struct {
		Location string  `json:"location"`
		GasType  string  `json:"gasType"`
		Value    float64 `json:"value"`
		Level    int     `json:"level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	alert := map[string]interface{}{
		"location": req.Location,
		"gasType":  req.GasType,
		"value":    req.Value,
		"level":    req.Level,
	}

	result, err := h.analyzer.AnalyzeAlert(c.Request.Context(), alert)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

func (h *AIHandler) Recommend(c *gin.Context) {
	var req struct {
		Location string  `json:"location"`
		GasType  string  `json:"gasType"`
		Value    float64 `json:"value"`
		Level    int     `json:"level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	alert := model.Alert{
		Location: req.Location,
		GasType:  req.GasType,
		Value:    req.Value,
		Level:    req.Level,
	}

	plans, err := h.recommender.Recommend(c.Request.Context(), alert)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"plans": plans})
}
