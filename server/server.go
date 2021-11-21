package server

import (
	"log"
	"net/http"

	"github.com/crossjoin-io/crossjoin/api"
)

// Server is an API server.
type Server struct {
	listenAddress string
	api           *api.API
}

// NewServer creates a server instance.
func NewServer(listenAddress, dataDir string) (*Server, error) {
	api := api.NewAPI(dataDir)
	return &Server{
		listenAddress: listenAddress,
		api:           api,
	}, nil
}

func (s *Server) Start() error {
	log.Printf("listening on %s", s.listenAddress)
	return http.ListenAndServe(s.listenAddress, s.api.Handler())
}
