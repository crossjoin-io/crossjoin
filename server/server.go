package server

import (
	"net/http"
)

// Server is an API server.
type Server struct {
	listenAddress string
}

// NewServer creates a server instance.
func NewServer(listenAddress string) (*Server, error) {
	return &Server{
		listenAddress: listenAddress,
	}, nil
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.listenAddress, nil)
}
