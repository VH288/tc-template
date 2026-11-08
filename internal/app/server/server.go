package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type HTTPServer struct {
	server *http.Server
	logger *slog.Logger
}

func NewHTTP(addr string, handler http.Handler, logger *slog.Logger, readTimeout, writeTimeout time.Duration) *HTTPServer {
	return &HTTPServer{
		logger: logger,
		server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
	}
}

func (s *HTTPServer) Start() error {
	s.logger.Info("http server starting", "addr", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("http server shutting down")
	return s.server.Shutdown(ctx)
}
