package eventbus

import (
	"testing"
	"time"
)

func TestEventBus_PublishSubscribe(t *testing.T) {
	eb := New()
	ch := make(chan Event, 10)
	eb.Subscribe("alert", ch)
	eb.Publish(Event{Type: "alert", Data: map[string]interface{}{"msg": "test"}})
	select {
	case e := <-ch:
		if e.Type != "alert" {
			t.Errorf("Type: got %s, want alert", e.Type)
		}
	case <-time.After(time.Second):
		t.Error("timeout: no event received")
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	eb := New()
	ch := make(chan Event, 10)
	eb.Subscribe("alert", ch)
	eb.Unsubscribe("alert", ch)
	eb.Publish(Event{Type: "alert", Data: nil})
	time.Sleep(100 * time.Millisecond)
	select {
	case <-ch:
		t.Error("should not receive after unsubscribe")
	default:
		// OK
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	eb := New()
	ch1 := make(chan Event, 10)
	ch2 := make(chan Event, 10)
	eb.Subscribe("alert", ch1)
	eb.Subscribe("alert", ch2)
	eb.Publish(Event{Type: "alert"})
	var count int
	for _, ch := range []chan Event{ch1, ch2} {
		select {
		case <-ch:
			count++
		case <-time.After(time.Second):
		}
	}
	if count != 2 {
		t.Errorf("subscribers: got %d, want 2", count)
	}
}

func TestEventBus_PublishToAll(t *testing.T) {
	eb := New()
	ch := make(chan Event, 10)
	eb.Subscribe("all", ch)
	eb.Publish(Event{Type: "alert"})
	select {
	case e := <-ch:
		if e.Type != "alert" {
			t.Errorf("Type: got %s, want alert", e.Type)
		}
	case <-time.After(time.Second):
		t.Error("timeout: no event received on 'all' subscriber")
	}
}

func TestEventBus_NonBlockingSend(t *testing.T) {
	eb := New()
	ch := make(chan Event, 1) // buffered size 1, will fill up
	eb.Subscribe("alert", ch)
	// Fill the channel
	ch <- Event{Type: "alert"}
	// Publish should not block (non-blocking)
	eb.Publish(Event{Type: "alert", Data: nil})
	// Original event should still be there
	select {
	case e := <-ch:
		if e.Type != "alert" {
			t.Errorf("Type: got %s, want alert", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout: event disappeared from full channel")
	}
}
