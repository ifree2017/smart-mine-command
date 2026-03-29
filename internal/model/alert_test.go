package model

import (
	"testing"
	"time"
)

func TestAlert_Fields(t *testing.T) {
	a := Alert{
		ID:       "test-id",
		Source:   "kj90x",
		Level:    3,
		Location: "001",
		GasType:  "T",
		Value:    1.5,
		Status:   "open",
	}
	if a.ID != "test-id" {
		t.Errorf("ID: got %s, want test-id", a.ID)
	}
	if a.Level != 3 {
		t.Errorf("Level: got %d, want 3", a.Level)
	}
	if a.Status != "open" {
		t.Errorf("Status: got %s, want open", a.Status)
	}
	if a.Source != "kj90x" {
		t.Errorf("Source: got %s, want kj90x", a.Source)
	}
	if a.GasType != "T" {
		t.Errorf("GasType: got %s, want T", a.GasType)
	}
	if a.Value != 1.5 {
		t.Errorf("Value: got %f, want 1.5", a.Value)
	}
}

func TestAlert_CreatedAt(t *testing.T) {
	now := time.Now()
	a := Alert{CreatedAt: now}
	if a.CreatedAt != now {
		t.Errorf("CreatedAt: got %v, want %v", a.CreatedAt, now)
	}
}

func TestCommand_Fields(t *testing.T) {
	c := Command{
		ID:      "cmd-1",
		Type:    "broadcast",
		Target:  "all",
		Content: "test message",
		Status:  "pending",
	}
	if c.Type != "broadcast" {
		t.Errorf("Type: got %s, want broadcast", c.Type)
	}
	if c.Target != "all" {
		t.Errorf("Target: got %s, want all", c.Target)
	}
	if c.Content != "test message" {
		t.Errorf("Content: got %s, want test message", c.Content)
	}
	if c.Status != "pending" {
		t.Errorf("Status: got %s, want pending", c.Status)
	}
}

func TestCommand_CreatedAt(t *testing.T) {
	now := time.Now()
	c := Command{CreatedAt: now}
	if c.CreatedAt != now {
		t.Errorf("CreatedAt: got %v, want %v", c.CreatedAt, now)
	}
}
