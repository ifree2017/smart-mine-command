package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"smart-mine-command/internal/ai"
	"smart-mine-command/internal/dispatch"
	"smart-mine-command/internal/eventbus"
	"smart-mine-command/internal/model"
	"smart-mine-command/internal/store"

	"smart-mine-command/internal/api/handler"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	engine     *gin.Engine
	eb         *eventbus.EventBus
	store      *store.Store
	disp       *dispatch.Dispatcher
	executor   *dispatch.CommandExecutor
	cmdHandler *handler.CommandHandler
	aiHandler  *handler.AIHandler
}

func NewServer(eb *eventbus.EventBus, s *store.Store, disp *dispatch.Dispatcher) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	executor := dispatch.NewCommandExecutor(eb)
	cmdHandler := handler.NewCommandHandler(executor, s)

	// 初始化AI组件
	aiClient := ai.NewClient("https://api.openai.com/v1", "")
	analyzer := ai.NewAnalyzer(aiClient)
	recommender := ai.NewRecommender(analyzer)
	aiHandler := handler.NewAIHandler(analyzer, recommender)

	srv := &Server{
		engine:     engine,
		eb:         eb,
		store:      s,
		disp:       disp,
		executor:   executor,
		cmdHandler: cmdHandler,
		aiHandler:  aiHandler,
	}

	engine.GET("/health", srv.handleHealth)
	engine.GET("/ws", srv.handleWS)

	api := engine.Group("/api")
	api.GET("/alerts", srv.handleListAlerts)
	api.POST("/alerts", srv.handleCreateAlert)
	api.POST("/alerts/ack/:id", srv.handleAckAlert)
	api.GET("/commands", srv.cmdHandler.List)
	api.POST("/commands", srv.cmdHandler.Create)
	api.GET("/commands/:id", srv.cmdHandler.Get)

	api.POST("/ai/analyze", srv.aiHandler.Analyze)
	api.POST("/ai/recommend", srv.aiHandler.Recommend)

	return srv
}

func (s *Server) Run(addr string) error {
	log.Printf("[api] server starting on %s", addr)
	return s.engine.Run(addr)
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "time": time.Now()})
}

func (s *Server) handleListAlerts(c *gin.Context) {
	alerts, err := s.store.ListAlerts(100)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, alerts)
}

func (s *Server) handleCreateAlert(c *gin.Context) {
	var a model.Alert
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	a.ID = uuid.New().String()
	a.Source = "manual"
	if a.Status == "" {
		a.Status = "open"
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	if err := s.store.SaveAlert(&a); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	s.eb.Publish(eventbus.Event{
		Type:   eventbus.EventTypeAlert,
		Source: "manual",
		Data:   map[string]interface{}{"alert": a},
	})
	c.JSON(200, a)
}

func (s *Server) handleAckAlert(c *gin.Context) {
	id := c.Param("id")
	err := s.store.UpdateAlert(id, map[string]interface{}{"status": "acknowledged"})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"id": id, "status": "acknowledged"})
}

func (s *Server) handleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[ws] upgrade error: %v", err)
		return
	}
	defer conn.Close()

	ch := make(chan eventbus.Event, 100)
	s.eb.Subscribe(eventbus.EventTypeAlert, ch)
	defer s.eb.Unsubscribe(eventbus.EventTypeAlert, ch)
	s.eb.Subscribe("all", ch)
	defer s.eb.Unsubscribe("all", ch)

	log.Printf("[ws] client connected")

	connClose := make(chan struct{})
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				close(connClose)
				return
			}
		}
	}()

	for {
		select {
		case e := <-ch:
			data, err := json.Marshal(e)
			if err != nil {
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("[ws] write error: %v", err)
				return
			}
		case <-connClose:
			log.Printf("[ws] client disconnected")
			return
		}
	}
}
