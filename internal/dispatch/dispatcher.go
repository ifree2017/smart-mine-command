package dispatch

import (
	"context"
	"fmt"
	"log"

	"smart-mine-command/internal/eventbus"
	"smart-mine-command/internal/model"
)

type Alert = model.Alert

type Dispatcher struct {
	eb *eventbus.EventBus
}

func NewDispatcher(eb *eventbus.EventBus) *Dispatcher {
	return &Dispatcher{eb: eb}
}

func (d *Dispatcher) HandleAlert(ctx context.Context, alert Alert) error {
	// 发布事件到 EventBus
	d.eb.Publish(eventbus.Event{
		Type:   eventbus.EventTypeAlert,
		Source: alert.Source,
		Data: map[string]interface{}{
			"alert": alert,
		},
	})

	// 如果超过阈值，触发广播指令
	if d.shouldBroadcast(alert) {
		d.broadcastCommand(alert)
	}

	log.Printf("[dispatch] handled alert: %s %s=%.2f level=%d", alert.Location, alert.GasType, alert.Value, alert.Level)
	return nil
}

func (d *Dispatcher) shouldBroadcast(alert Alert) bool {
	// 5级报警或3级以上高危气体
	return alert.Level >= 4
}

func (d *Dispatcher) broadcastCommand(alert Alert) {
	cmd := model.Command{
		Type:    "broadcast",
		Target:  "all",
		Content: "紧急报警：" + alert.Location + " " + alert.GasType + "浓度超标(" + fmt.Sprintf("%.2f", alert.Value) + ")",
		Status:  "pending",
	}
	d.eb.Publish(eventbus.Event{
		Type:   eventbus.EventTypeCommand,
		Source: "dispatcher",
		Data:   map[string]interface{}{"command": cmd},
	})
}
