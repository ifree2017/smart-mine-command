package store

import (
	"os"
	"testing"

	"smart-mine-command/internal/model"
)

func TestStore_SaveAndListAlert(t *testing.T) {
	tmp := t.TempDir() + "/store.json"
	s := NewStore(tmp)
	a := model.Alert{Level: 2}
	if err := s.SaveAlert(&a); err != nil {
		t.Fatalf("SaveAlert: %v", err)
	}
	alerts, err := s.ListAlerts(10)
	if err != nil {
		t.Fatalf("ListAlerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("len: got %d, want 1", len(alerts))
	}
}

func TestStore_SaveAndListCommand(t *testing.T) {
	tmp := t.TempDir() + "/store.json"
	s := NewStore(tmp)
	c := model.Command{Type: "broadcast"}
	if err := s.SaveCommand(&c); err != nil {
		t.Fatalf("SaveCommand: %v", err)
	}
	cmds, err := s.ListCommands()
	if err != nil {
		t.Fatalf("ListCommands: %v", err)
	}
	if len(cmds) != 1 {
		t.Errorf("len: got %d, want 1", len(cmds))
	}
}

func TestStore_UpdateAlert(t *testing.T) {
	tmp := t.TempDir() + "/store.json"
	s := NewStore(tmp)
	s.alerts = []model.Alert{{ID: "a1", Status: "open"}}
	err := s.UpdateAlert("a1", map[string]interface{}{"status": "acknowledged"})
	if err != nil {
		t.Fatalf("UpdateAlert: %v", err)
	}
	if s.alerts[0].Status != "acknowledged" {
		t.Errorf("Status: got %s, want acknowledged", s.alerts[0].Status)
	}
}

func TestStore_ListAlerts_Limit(t *testing.T) {
	tmp := t.TempDir() + "/store.json"
	s := NewStore(tmp)
	for i := 0; i < 5; i++ {
		s.SaveAlert(&model.Alert{Level: 1})
	}
	alerts, err := s.ListAlerts(3)
	if err != nil {
		t.Fatalf("ListAlerts: %v", err)
	}
	if len(alerts) != 3 {
		t.Errorf("len: got %d, want 3", len(alerts))
	}
}

func TestStore_ListAlerts_All(t *testing.T) {
	tmp := t.TempDir() + "/store.json"
	s := NewStore(tmp)
	for i := 0; i < 5; i++ {
		s.SaveAlert(&model.Alert{Level: 1})
	}
	alerts, err := s.ListAlerts(0) // 0 means all
	if err != nil {
		t.Fatalf("ListAlerts: %v", err)
	}
	if len(alerts) != 5 {
		t.Errorf("len: got %d, want 5", len(alerts))
	}
}

func TestStore_Persist(t *testing.T) {
	tmp := t.TempDir() + "/store.json"
	s := NewStore(tmp)
	s.SaveAlert(&model.Alert{Level: 2})
	if err := s.Persist(); err != nil {
		t.Fatalf("Persist: %v", err)
	}
	// Read the file directly
	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Error("persist wrote empty file")
	}
}
