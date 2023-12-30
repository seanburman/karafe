package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/gommon/log"
)

var servers = make(map[int]*Server)

type (
	Server struct {
		http *http.Server
		Port int
		Path string
	}
	ServerConfig struct {
		Port       int
		Path       string
		Middleware []http.HandlerFunc
		Group      func(*http.ServeMux, ServerConfig) error
	}
)

func NewServer(cfg ServerConfig) (*Server, error) {
	for port, sv := range servers {
		if port == cfg.Port {
			return nil, fmt.Errorf("server with port %d already exists", cfg.Port)
		}
		if sv.Path == cfg.Path {
			return nil, fmt.Errorf("server with path %s already exists", cfg.Path)
		}
	}

	var addr string
	if cfg.Port <= 0 {
		addr = ":8080"
	} else {
		addr = fmt.Sprint(":", cfg.Port)
	}

	mux := http.NewServeMux()

	server := http.Server{
		Addr:    "localhost" + addr,
		Handler: mux,
	}

	ss := &Server{
		&server,
		cfg.Port,
		cfg.Path,
	}

	go func() {
		log.Infof("serving http://%d/%s", cfg.Path, server.Addr)
		server.ListenAndServe()
	}()
	servers[ss.Port] = ss

	return ss, nil
}

func (s *Server) Close() error {
	delete(servers, s.Port)
	return s.http.Close()
}
