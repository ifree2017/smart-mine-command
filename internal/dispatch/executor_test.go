package dispatch

import (
	"context"
	"testing"

	"smart-mine-command/internal/eventbus"
	"smart-mine-command/internal/model"
)

func TestCommandExecutor_Execute_Broadcast(t *testing.T) {
	eb := eventbus.New()
	executor := NewCommandExecutor(eb)

	cmd := &model.Command{ID: "c1", Type: "broadcast", Target: "all", Content: "test"}
	err := executor.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if cmd.Status != "done" {
		t.Errorf("Status: got %s, want done", cmd.Status)
	}
}

func TestCommandExecutor_Execute_WS(t *testing.T) {
	eb := eventbus.New()
	executor := NewCommandExecutor(eb)

	cmd := &model.Command{ID: "c2", Type: "ws", Target: "group1", Content: "ws test"}
	err := executor.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if cmd.Status != "done" {
		t.Errorf("Status: got %s, want done", cmd.Status)
	}
}

func TestCommandExecutor_Execute_Call(t *testing.T) {
	eb := eventbus.New()
	executor := NewCommandExecutor(eb)

	cmd := &model.Command{ID: "c3", Type: "call", Target: "person1", Content: "call test"}
	err := executor.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if cmd.Status != "done" {
		t.Errorf("Status: got %s, want done", cmd.Status)
	}
}

func TestCommandExecutor_Execute_Unknown(t *testing.T) {
	eb := eventbus.New()
	executor := NewCommandExecutor(eb)

	cmd := &model.Command{ID: "c4", Type: "unknown", Content: "unknown"}
	err := executor.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if cmd.Status != "failed" {
		t.Errorf("Status: got %s, want failed", cmd.Status)
	}
}
