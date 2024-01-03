package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"nhooyr.io/websocket"
)

var manager = &serverManager{
	servers: make(map[string]*Server),
}

type (
	serverManager struct {
		mu      sync.Mutex
		servers map[string]*Server
	}
	Server struct {
		app    *echo.Echo
		Config ServerConfig

		subscribers map[*subscriber]bool
		subMu       sync.Mutex
	}
	Route struct {
		Path string
	}
	ServerConfig struct {
		Name string
		Port string
		Path string
	}
	subscriber struct {
		msgs chan []byte
	}
)

func NewServer(cfg ServerConfig) (*Server, error) {
	manager.mu.Lock()
	for port, sv := range manager.servers {
		if port == cfg.Port {
			return nil, fmt.Errorf("server with port %s already exists", cfg.Port)
		}
		if sv.cfg.Path == cfg.Path {
			return nil, fmt.Errorf("server with path %s already exists", cfg.Path)
		}
	}

	server := &Server{
		app:         echo.New(),
		Config:      cfg,
		subscribers: make(map[*subscriber]bool),
	}

	manager.servers[server.Config.Port] = server
	manager.mu.Unlock()

	server.app.Static(cfg.Path+"/client", "public")
	ws := server.app.Group(cfg.Path + "/ws")
	ws.GET("/subscribe", server.handleSubscribe)
	return server, nil
}

func (s *Server) ListenAndServe() {
	go s.app.Start(s.Config.Port)
}

func (s *Server) Shutdown() {
	manager.mu.Lock()
	delete(manager.servers, s.Config.Port)
	manager.mu.Unlock()
}

func (s *Server) handleSubscribe(ctx echo.Context) error {

	return nil
}

func (s *Server) subscribe(ctx echo.Context) error {
	var mu sync.Mutex
	var conn *websocket.Conn

	sub := &subscriber{
		msgs: make(chan []byte, 16),
	}
	s.addSubscriber(sub)
	defer s.deleteSubscriber(sub)

	socket, err := websocket.Accept(ctx.Response().Writer, ctx.Request(), nil)
	if err != nil {
		return err
	}
	mu.Lock()
	conn = socket
	mu.Unlock()
	defer conn.CloseNow()

	context := conn.CloseRead(ctx.Request().Context())

	for {
		select {
		case msg := <-sub.msgs:
			err := writeTimeout(context, time.Second*5, conn, msg)
			if err != nil {
				return err
			}
		case <-context.Done():
			return context.Err()
		}
	}
}

func (s *Server) publish(msg []byte) {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	for sub := range s.subscribers {
		select {
		case sub.msgs <- msg:
		default:
			continue
		}
	}
}

func (s *Server) addSubscriber(sub *subscriber) {
	s.subMu.Lock()
	s.subscribers[sub] = true
	s.subMu.Unlock()
}

func (s *Server) deleteSubscriber(sub *subscriber) {
	s.subMu.Lock()
	delete(s.subscribers, sub)
	s.subMu.Unlock()
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}
