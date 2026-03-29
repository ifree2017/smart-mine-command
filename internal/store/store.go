package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"smart-mine-command/internal/model"
)

type Store struct {
	alerts   []model.Alert
	commands []model.Command
	mu       sync.RWMutex
	file     string
}

func NewStore(path string) *Store {
	s := &Store{
		alerts:   []model.Alert{},
		commands: []model.Command{},
		file:     path,
	}
	s.load()
	return s
}

func (s *Store) load() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return
	}
	var saved struct {
		Alerts   []model.Alert   `json:"alerts"`
		Commands []model.Command `json:"commands"`
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		return
	}
	s.alerts = saved.Alerts
	s.commands = saved.Commands
}

func (s *Store) SaveAlert(a *model.Alert) error {
	if a.ID == "" {
		a.ID = fmt.Sprintf("ALT-%d", time.Now().UnixNano())
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	if a.Status == "" {
		a.Status = "open"
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts = append(s.alerts, *a)
	return s.persistLocked()
}

func (s *Store) ListAlerts(limit int) ([]model.Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > len(s.alerts) {
		limit = len(s.alerts)
	}
	alerts := make([]model.Alert, limit)
	copy(alerts, s.alerts[len(s.alerts)-limit:])
	return alerts, nil
}

func (s *Store) UpdateAlert(id string, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.alerts {
		if s.alerts[i].ID == id {
			if status, ok := updates["status"].(string); ok {
				s.alerts[i].Status = status
			}
			break
		}
	}
	return s.persistLocked()
}

func (s *Store) SaveCommand(c *model.Command) error {
	if c.ID == "" {
		c.ID = fmt.Sprintf("CMD-%d", time.Now().UnixNano())
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	if c.Status == "" {
		c.Status = "pending"
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.commands = append(s.commands, *c)
	return s.persistLocked()
}

func (s *Store) ListCommands() ([]model.Command, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.commands, nil
}

func (s *Store) Persist() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.persistLocked()
}

func (s *Store) persistLocked() error {
	if s.file == "" {
		return nil
	}
	data, err := json.MarshalIndent(struct {
		Alerts   []model.Alert   `json:"alerts"`
		Commands []model.Command `json:"commands"`
	}{s.alerts, s.commands}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.file, data, 0644)
}
