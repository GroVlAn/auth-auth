package http_server

import (
	"context"
	"net/http"
	"time"
)

type Settings struct {
	Port              string
	MaxHeaderBytes    int
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
}

type Server struct {
	httpServer *http.Server
}

func New(handler http.Handler, setting Settings) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + setting.Port,
			Handler:           handler,
			MaxHeaderBytes:    setting.MaxHeaderBytes,
			ReadHeaderTimeout: setting.ReadHeaderTimeout,
			WriteTimeout:      setting.WriteTimeout,
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
