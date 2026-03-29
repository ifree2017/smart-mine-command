package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"smart-mine-command/internal/dispatch"
	"smart-mine-command/internal/eventbus"
	"smart-mine-command/internal/model"
	"smart-mine-command/internal/store"
)

func TestCommandHandler_Create(t *testing.T) {
	tmp := t.TempDir() + "/store.json"
	s := store.NewStore(tmp)
	eb := eventbus.New()
	executor := dispatch.NewCommandExecutor(eb)
	h := NewCommandHandler(executor, s)

	req := CreateCommandRequest{
		Type:    "broadcast",
		Target:  "all",
		Content: "test broadcast",
	}
	body, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/commands", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create(c)

	if w.Code != 200 {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

func TestCommandHandler_List(t *testing.T) {
	tmp := t.TempDir() + "/store.json"
	s := store.NewStore(tmp)
	eb := eventbus.New()
	executor := dispatch.NewCommandExecutor(eb)
	h := NewCommandHandler(executor, s)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/commands", nil)

	h.List(c)

	if w.Code != 200 {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

func TestCommandHandler_Get(t *testing.T) {
	tmp := t.TempDir() + "/store.json"
	s := store.NewStore(tmp)
	eb := eventbus.New()
	executor := dispatch.NewCommandExecutor(eb)
	h := NewCommandHandler(executor, s)

	// 先创建一个 command
	s.SaveCommand(&model.Command{ID: "test-cmd", Type: "broadcast"})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/commands/test-cmd", nil)
	c.Params = []gin.Param{{Key: "id", Value: "test-cmd"}}

	h.Get(c)

	if w.Code != 200 {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}

func TestCommandHandler_Get_NotFound(t *testing.T) {
	tmp := t.TempDir() + "/store.json"
	s := store.NewStore(tmp)
	eb := eventbus.New()
	executor := dispatch.NewCommandExecutor(eb)
	h := NewCommandHandler(executor, s)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/commands/not-exist", nil)
	c.Params = []gin.Param{{Key: "id", Value: "not-exist"}}

	h.Get(c)

	if w.Code != 404 {
		t.Errorf("status: got %d, want 404", w.Code)
	}
}
