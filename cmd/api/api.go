package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/gommon/log"
	"github.com/rs/cors"
)

var servers = make(map[string]*Server)

type (
	Server struct {
		//TODO:
		// connections map[*websocket.Conn]true
		Config ServerConfig
		*http.ServeMux
		Ctx        context.Context
		middleware []func(http.Handler) http.Handler
		shutdown   func(context.Context) error
	}
	Route struct {
		Path       string
		middleware []func(http.Handler) http.Handler
		Handler    func(http.ResponseWriter, *http.Request)
	}
	ServerConfig struct {
		Port   string
		Path   string
		Routes []*Route
	}
)

func NewServer(cfg ServerConfig) (*Server, error) {
	for port, sv := range servers {
		if port == cfg.Port {
			return nil, fmt.Errorf("server with port %s already exists", cfg.Port)
		}
		if sv.Config.Path == cfg.Path {
			return nil, fmt.Errorf("server with path %s already exists", cfg.Path)
		}
	}

	server := &Server{
		ServeMux: http.NewServeMux(),
		Config:   cfg,
		Ctx:      context.Background(),
	}

	servers[server.Config.Port] = server

	return server, nil
}

func (s *Server) ListenAndServe() {
	s.ServeMux.Handle("/", http.FileServer(http.Dir("./build")))
	s.applyMiddleware()
	CORS := cors.Default().Handler(s.ServeMux)
	srv := &http.Server{
		Addr:    "localhost" + s.Config.Port,
		Handler: CORS,
	}
	s.shutdown = srv.Shutdown
	log.Infof("serving http://%s%s", srv.Addr, s.Config.Path)
	go srv.ListenAndServe()
}

func (s *Server) Shutdown() {
	err := s.shutdown(s.Ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) UseRoute(r *Route) {
	s.Config.Routes = append(s.Config.Routes, r)
}

func (s *Server) UseMiddleware(middleware func(http.Handler) http.Handler) {
	mw := []func(http.Handler) http.Handler{middleware}
	mw = append(mw, middleware)
	s.middleware = mw
}

func (s *Server) applyMiddleware() {
	for _, route := range s.Config.Routes {
		var handler http.Handler
		for k, middleware := range route.middleware {
			if k == 0 {
				handler = http.HandlerFunc(route.Handler)
				continue
			}
			handler = middleware(handler)
		}
		for _, middleware := range s.middleware {
			handler = middleware(handler)
		}
		if len(route.middleware) == 0 && len(s.middleware) == 0 {
			handler = http.HandlerFunc(route.Handler)
		}
		s.ServeMux.Handle(s.Config.Path+route.Path, handler)
	}
}

func NewRoute(path string) *Route {
	return &Route{}
}

func (r *Route) HandleFunc(path string, handler http.HandlerFunc) {
	r.Path = path
	r.Handler = handler
}

func (r *Route) UseMiddleware(middleware func(http.Handler) http.Handler) {
	mw := []func(http.Handler) http.Handler{middleware}
	mw = append(mw, middleware)
	r.middleware = mw
}