package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"smart-mine-command/internal/dispatch"
	"smart-mine-command/internal/model"
	"smart-mine-command/internal/store"
)

type CommandHandler struct {
	executor *dispatch.CommandExecutor
	store    *store.Store
}

func NewCommandHandler(executor *dispatch.CommandExecutor, s *store.Store) *CommandHandler {
	return &CommandHandler{executor: executor, store: s}
}

type CreateCommandRequest struct {
	Type    string `json:"type"`    // broadcast/ws/call
	Target  string `json:"target"`  // 群组ID/设备ID/人员ID
	Content string `json:"content"` // 指令内容
}

func (h *CommandHandler) Create(c *gin.Context) {
	var req CreateCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	cmd := &model.Command{
		ID:      uuid.New().String(),
		Type:    req.Type,
		Target:  req.Target,
		Content: req.Content,
		Status:  "pending",
	}

	if err := h.store.SaveCommand(cmd); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 异步执行
	go h.executor.Execute(c.Request.Context(), cmd)

	c.JSON(200, gin.H{"command": cmd})
}

func (h *CommandHandler) List(c *gin.Context) {
	cmds, err := h.store.ListCommands()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"commands": cmds})
}

func (h *CommandHandler) Get(c *gin.Context) {
	id := c.Param("id")
	cmds, err := h.store.ListCommands()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	for _, cmd := range cmds {
		if cmd.ID == id {
			c.JSON(200, gin.H{"command": cmd})
			return
		}
	}
	c.JSON(404, gin.H{"error": "not found"})
}
