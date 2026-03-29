package ai

import (
	"context"
	"fmt"

	"smart-mine-command/internal/dispatch"
	"smart-mine-command/internal/model"
)

// Recommender 预案推荐器
type Recommender struct {
	analyzer *Analyzer
}

// NewRecommender 创建推荐器
func NewRecommender(analyzer *Analyzer) *Recommender {
	return &Recommender{analyzer: analyzer}
}

// Recommend 推荐预案
func (r *Recommender) Recommend(ctx context.Context, alert model.Alert) ([]dispatch.Plan, error) {
	// 规则匹配 + AI辅助
	cmds := dispatch.Evaluate(alert, dispatch.DefaultRules)

	// 构建预案
	plans := r.buildPlans(alert, cmds)

	// AI增强：如果等级>=3，用AI补充建议
	if alert.Level >= 3 {
		alertMap := map[string]interface{}{
			"location": alert.Location,
			"gasType":  alert.GasType,
			"value":    alert.Value,
			"level":    alert.Level,
		}
		result, err := r.analyzer.AnalyzeAlert(ctx, alertMap)
		if err == nil && result != nil {
			// 用AI结果更新预案建议
			plans = r.enrichWithAI(plans, result)
		}
	}

	return plans, nil
}

func (r *Recommender) buildPlans(alert model.Alert, cmds []model.Command) []dispatch.Plan {
	if len(cmds) == 0 {
		return nil
	}

	plan := dispatch.Plan{
		ID:       fmt.Sprintf("plan-%d", alert.Level),
		Name:     fmt.Sprintf("L%d预案", alert.Level),
		Trigger:  fmt.Sprintf("level>=%d", alert.Level),
		Commands: cmds,
	}

	return []dispatch.Plan{plan}
}

func (r *Recommender) enrichWithAI(plans []dispatch.Plan, result *AnalysisResult) []dispatch.Plan {
	for i := range plans {
		plans[i].Name = fmt.Sprintf("%s | AI建议:%s", plans[i].Name, result.Recommendation)
	}
	return plans
}
