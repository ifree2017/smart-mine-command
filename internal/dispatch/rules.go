package dispatch

import (
	"strings"

	"smart-mine-command/internal/model"
)

// Rule 调度规则
type Rule struct {
	LevelMin  int    // 最小报警等级
	LevelMax  int    // 最大报警等级
	GasType   string // 气体类型（空=全部）
	Action    string // broadcast/ws/call
	Template  string // 指令模板
}

// DefaultRules 默认调度规则
var DefaultRules = []Rule{
	{LevelMin: 4, Action: "broadcast", Template: "紧急疏散：{location}发生{gasType}浓度超标"},
	{LevelMin: 3, Action: "ws", Template: "注意：{location}{gasType}异常"},
	{LevelMin: 1, Action: "ws", Template: "通知：{location}{gasType}监测"},
}

// Evaluate 评估报警，返回应执行的指令
func Evaluate(alert model.Alert, rules []Rule) []model.Command {
	var cmds []model.Command
	for _, rule := range rules {
		if alert.Level >= rule.LevelMin && (rule.LevelMax == 0 || alert.Level <= rule.LevelMax) {
			if rule.GasType == "" || rule.GasType == alert.GasType {
				cmds = append(cmds, model.Command{
					Type:    rule.Action,
					Target:  "all",
					Content: fillTemplate(rule.Template, alert),
					Status:  "pending",
				})
			}
		}
	}
	return cmds
}

func fillTemplate(template string, alert model.Alert) string {
	result := template
	result = strings.ReplaceAll(result, "{location}", alert.Location)
	result = strings.ReplaceAll(result, "{gasType}", alert.GasType)
	return result
}
