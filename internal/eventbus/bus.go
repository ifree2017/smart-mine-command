package eventbus

import (
	"fmt"
	"sync"
	"time"
)

const (
	EventTypeAlert    = "alert"
	EventTypeDispatch = "dispatch"
	EventTypeCommand  = "command"
	EventTypePlan     = "plan"
)

type Event struct {
	Type    string                 `json:"type"`    // "alert"|"dispatch"|"command"|"plan"
	Source  string                 `json:"source"`  // "kj90x"|"manual"|"ai"
	Data    map[string]interface{} `json:"data"`
	TraceID string                 `json:"trace_id"`
	Time    time.Time              `json:"time"`
}

type EventBus struct {
	subscribers map[string][]chan Event
	mu          sync.RWMutex
}

func New() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
	}
}

func (eb *EventBus) Subscribe(topic string, ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers[topic] = append(eb.subscribers[topic], ch)
}

func (eb *EventBus) Unsubscribe(topic string, ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	chans := eb.subscribers[topic]
	for i, c := range chans {
		if c == ch {
			eb.subscribers[topic] = append(chans[:i], chans[i+1:]...)
			return
		}
	}
}

func (eb *EventBus) Publish(e Event) {
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	if e.TraceID == "" {
		e.TraceID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	eb.mu.RLock()
	defer eb.mu.RUnlock()

	for _, ch := range eb.subscribers[e.Type] {
		select {
		case ch <- e:
		default:
			// non-blocking send
		}
	}
	// also publish to "all" topic
	for _, ch := range eb.subscribers["all"] {
		select {
		case ch <- e:
		default:
		}
	}
}
