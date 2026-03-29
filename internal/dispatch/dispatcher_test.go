package dispatch

import (
	"context"
	"testing"
	"time"

	"smart-mine-command/internal/eventbus"
	"smart-mine-command/internal/model"
)

func TestDispatcher_HandleAlert(t *testing.T) {
	eb := eventbus.New()
	d := NewDispatcher(eb)
	alert := model.Alert{ID: "a1", Location: "001", Level: 2, Source: "test"}
	err := d.HandleAlert(context.Background(), alert)
	if err != nil {
		t.Errorf("HandleAlert: %v", err)
	}
}

func TestDispatcher_HandleAlert_Broadcasts(t *testing.T) {
	eb := eventbus.New()
	d := NewDispatcher(eb)
	ch := make(chan eventbus.Event, 10)
	eb.Subscribe("alert", ch)
	d.HandleAlert(context.Background(), model.Alert{ID: "a1", Location: "001", Level: 2, Source: "test"})
	select {
	case e := <-ch:
		if e.Type != "alert" {
			t.Errorf("Type: got %s, want alert", e.Type)
		}
	case <-time.After(time.Second):
		t.Error("timeout")
	}
}

func TestDispatcher_ShouldBroadcast_HighLevel(t *testing.T) {
	eb := eventbus.New()
	d := NewDispatcher(eb)
	// Level 4 should trigger broadcast
	cmdCh := make(chan eventbus.Event, 10)
	eb.Subscribe("command", cmdCh)
	d.HandleAlert(context.Background(), model.Alert{ID: "a1", Location: "001", Level: 4, Source: "test", GasType: "T", Value: 2.0})
	select {
	case e := <-cmdCh:
		if e.Type != "command" {
			t.Errorf("Type: got %s, want command", e.Type)
		}
	case <-time.After(time.Second):
		t.Error("timeout: no command broadcast for high-level alert")
	}
}

func TestDispatcher_ShouldBroadcast_Level5(t *testing.T) {
	eb := eventbus.New()
	d := NewDispatcher(eb)
	cmdCh := make(chan eventbus.Event, 10)
	eb.Subscribe("command", cmdCh)
	d.HandleAlert(context.Background(), model.Alert{ID: "a1", Location: "002", Level: 5, Source: "test", GasType: "CO", Value: 100.0})
	select {
	case <-cmdCh:
		// OK - level 5 should trigger broadcast
	case <-time.After(time.Second):
		t.Error("timeout: no command broadcast for level-5 alert")
	}
}

func TestDispatcher_NoBroadcast_LowLevel(t *testing.T) {
	eb := eventbus.New()
	d := NewDispatcher(eb)
	cmdCh := make(chan eventbus.Event, 10)
	eb.Subscribe("command", cmdCh)
	d.HandleAlert(context.Background(), model.Alert{ID: "a1", Location: "001", Level: 2, Source: "test"})
	// Wait briefly then check no command was sent
	alert := func() bool {
		select {
		case <-cmdCh:
			return true
		case <-time.After(200 * time.Millisecond):
			return false
		}
	}()
	if alert {
		t.Error("should not broadcast for low-level alert")
	}
}
