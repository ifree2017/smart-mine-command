package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"smart-mine-command/internal/api"
	"smart-mine-command/internal/dispatch"
	"smart-mine-command/internal/eventbus"
	"smart-mine-command/internal/model"
	"smart-mine-command/internal/protocols/kj90x"
	"smart-mine-command/internal/store"
)

func main() {
	eb := eventbus.New()
	s := store.NewStore("data.json")
	disp := dispatch.NewDispatcher(eb)

	// Subscribe to alerts and save to store
	alertCh := make(chan eventbus.Event, 100)
	eb.Subscribe(eventbus.EventTypeAlert, alertCh)
	go func() {
		for e := range alertCh {
			if alertData, ok := e.Data["alert"]; ok {
				if a, ok := alertData.(model.Alert); ok {
					s.SaveAlert(&a)
				}
			}
		}
	}()

	// TCP server for kj90x devices
	kj90xServer := kj90x.NewServer(":54321", eb)
	if err := kj90xServer.Run(); err != nil {
		log.Printf("[kj90x] server failed: %v", err)
	} else {
		log.Printf("[kj90x] server running on :54321")
	}

	// HTTP API + WebSocket
	httpServer := api.NewServer(eb, s, disp)
	go func() {
		if err := httpServer.Run(":8080"); err != nil {
			log.Printf("[api] server error: %v", err)
		}
	}()

	log.Println("[main] smart-mine-command started")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("[main] shutting down...")
	kj90xServer.Stop()
	s.Persist()
	log.Println("[main] done")
}
