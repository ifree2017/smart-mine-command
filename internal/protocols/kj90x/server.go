package kj90x

import (
	"bufio"
	"context"
	"log"
	"net"
	"sync"

	"smart-mine-command/internal/eventbus"
)

type Server struct {
	addr string
	eb   *eventbus.EventBus
	ln   net.Listener
	wg   sync.WaitGroup
	ctx  context.Context
	cancel context.CancelFunc
}

func NewServer(addr string, eb *eventbus.EventBus) *Server {
	return &Server{addr: addr, eb: eb}
}

func (s *Server) Run() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = ln
	s.ctx, s.cancel = context.WithCancel(context.Background())
	log.Printf("[kj90x] TCP server listening on %s", s.addr)

	s.wg.Add(1)
	go s.acceptLoop()
	return nil
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				log.Printf("[kj90x] accept error: %v", err)
				continue
			}
		}
		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		alert, err := Parse(line)
		if err != nil {
			log.Printf("[kj90x] parse error: %v", err)
			continue
		}
		modelAlert := alert.ToModel()
		s.eb.Publish(eventbus.Event{
			Type:   eventbus.EventTypeAlert,
			Source: "kj90x",
			Data:   map[string]interface{}{"alert": modelAlert},
		})
		log.Printf("[kj90x] alert published: %s %s=%.2f level=%d", modelAlert.Location, modelAlert.GasType, modelAlert.Value, modelAlert.Level)
	}
}

func (s *Server) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.ln != nil {
		s.ln.Close()
	}
	s.wg.Wait()
}
