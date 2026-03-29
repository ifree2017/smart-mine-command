package dispatch

import (
	"context"
	"testing"
	"time"

	"smart-mine-command/internal/eventbus"
	"smart-mine-command/internal/model"
)

func TestPlanExecutor_Execute(t *testing.T) {
	eb := eventbus.New()
	executor := NewPlanExecutor(eb)

	ch := make(chan eventbus.Event, 10)
	eb.Subscribe(string(eventbus.EventTypePlan), ch)

	plan := Plan{
		ID:   "p1",
		Name: "test plan",
		Commands: []model.Command{
			{ID: "c1", Type: "broadcast", Content: "test"},
		},
	}
	alert := model.Alert{ID: "a1", Level: 4}

	err := executor.Execute(context.Background(), plan, alert)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	select {
	case e := <-ch:
		if e.Type != eventbus.EventTypePlan {
			t.Errorf("Type: got %s, want plan", e.Type)
		}
	case <-time.After(time.Second):
		t.Error("timeout")
	}
}
