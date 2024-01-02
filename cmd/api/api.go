package api

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/seanburman/kaw/handlers"
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
		app *fiber.App
		ServerConfig
	}
	Route struct {
		Path string
	}
	ServerConfig struct {
		Name string
		Port string
		Path string
	}
)

func NewServer(cfg ServerConfig) (*Server, error) {
	manager.mu.Lock()
	for port, sv := range manager.servers {
		if port == cfg.Port {
			return nil, fmt.Errorf("server with port %s already exists", cfg.Port)
		}
		if sv.Path == cfg.Path {
			return nil, fmt.Errorf("server with path %s already exists", cfg.Path)
		}
	}

	server := &Server{
		fiber.New(fiber.Config{
			AppName: "KRAW",
		}),
		cfg,
	}

	manager.servers[server.Port] = server
	manager.mu.Unlock()

	server.app.Static("", "public")
	server.app.Route(cfg.Path, func(api fiber.Router) {
		api.Get("/ws", handlers.HandleGetWebSocket)
	})

	return server, nil
}

func (s *Server) ListenAndServe() {
	go s.app.Listen(s.Port)
}

func (s *Server) Shutdown() {
	manager.mu.Lock()
	delete(manager.servers, s.Port)
	manager.mu.Unlock()
}
