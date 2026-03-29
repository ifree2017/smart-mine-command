package dispatch

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"smart-mine-command/internal/eventbus"
	"smart-mine-command/internal/model"
)

// CommandExecutor 指令执行器
type CommandExecutor struct {
	eb *eventbus.EventBus
}

// NewCommandExecutor 创建执行器
func NewCommandExecutor(eb *eventbus.EventBus) *CommandExecutor {
	return &CommandExecutor{eb: eb}
}

// Execute 执行指令
func (e *CommandExecutor) Execute(ctx context.Context, cmd *model.Command) error {
	cmd.Status = "executing"
	cmd.CreatedAt = time.Now()

	// 发布执行开始事件
	e.eb.Publish(eventbus.Event{
		Type:    eventbus.EventTypeCommand,
		Source:  "system",
		TraceID: cmd.ID,
		Data:    cmdToMap(cmd),
		Time:    time.Now(),
	})

	switch cmd.Type {
	case "broadcast":
		return e.executeBroadcast(ctx, cmd)
	case "ws":
		return e.executeWS(ctx, cmd)
	case "call":
		return e.executeCall(ctx, cmd)
	default:
		cmd.Status = "failed"
		return nil
	}
}

func (e *CommandExecutor) executeBroadcast(ctx context.Context, cmd *model.Command) error {
	// 模拟广播：打印到日志 + 发布完成事件
	log.Printf("[executor] broadcast executed: %s", cmd.Content)
	cmd.Status = "done"
	e.eb.Publish(eventbus.Event{
		Type:    "command_done",
		Source:  "system",
		TraceID: cmd.ID,
		Data:    map[string]interface{}{"status": "done", "type": "broadcast"},
		Time:    time.Now(),
	})
	return nil
}

func (e *CommandExecutor) executeWS(ctx context.Context, cmd *model.Command) error {
	// WebSocket 推送（通过 EventBus 分发）
	log.Printf("[executor] ws push: target=%s, content=%s", cmd.Target, cmd.Content)
	cmd.Status = "done"
	e.eb.Publish(eventbus.Event{
		Type:    "command_done",
		Source:  "system",
		TraceID: cmd.ID,
		Data:    map[string]interface{}{"status": "done", "type": "ws", "target": cmd.Target},
		Time:    time.Now(),
	})
	return nil
}

func (e *CommandExecutor) executeCall(ctx context.Context, cmd *model.Command) error {
	// 模拟电话呼叫
	log.Printf("[executor] call: target=%s, content=%s", cmd.Target, cmd.Content)
	cmd.Status = "done"
	e.eb.Publish(eventbus.Event{
		Type:    "command_done",
		Source:  "system",
		TraceID: cmd.ID,
		Data:    map[string]interface{}{"status": "done", "type": "call", "target": cmd.Target},
		Time:    time.Now(),
	})
	return nil
}

func cmdToMap(cmd *model.Command) map[string]interface{} {
	data, _ := json.Marshal(cmd)
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	return m
}
