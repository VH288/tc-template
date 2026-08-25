package server

import (
	"log/slog"
	"net/http"
)

type HTTPServer struct {
	server *http.Server
	logger *slog.Logger
}
