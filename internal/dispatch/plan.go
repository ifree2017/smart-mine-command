package dispatch

import (
	"context"

	"smart-mine-command/internal/eventbus"
	"smart-mine-command/internal/model"
)

// PlanExecutor 数字预案执行器
type PlanExecutor struct {
	eb *eventbus.EventBus
}

func NewPlanExecutor(eb *eventbus.EventBus) *PlanExecutor {
	return &PlanExecutor{eb: eb}
}

// Plan 预案
type Plan struct {
	ID       string
	Name     string
	Trigger  string // 触发条件，如 "level>=4"
	Commands []model.Command
}

// Execute 执行预案
func (p *PlanExecutor) Execute(ctx context.Context, plan Plan, alert model.Alert) error {
	// 发布预案执行事件
	p.eb.Publish(eventbus.Event{
		Type:   eventbus.EventTypePlan,
		Source: "system",
		Data: map[string]interface{}{
			"plan_id":   plan.ID,
			"plan_name": plan.Name,
			"alert_id":  alert.ID,
		},
	})

	// 按顺序执行指令
	for _, cmd := range plan.Commands {
		if err := p.executeCommand(ctx, &cmd); err != nil {
			return err
		}
	}
	return nil
}

func (p *PlanExecutor) executeCommand(ctx context.Context, cmd *model.Command) error {
	// 简化：直接发布事件，由 CommandExecutor 处理
	p.eb.Publish(eventbus.Event{
		Type:   eventbus.EventTypeCommand,
		Source: "plan",
		Data:   map[string]interface{}{"command": cmd},
	})
	return nil
}
