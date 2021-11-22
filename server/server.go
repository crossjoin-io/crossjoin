package server

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/crossjoin-io/crossjoin/api"
	_ "github.com/mattn/go-sqlite3"
)

// Server is an API server.
type Server struct {
	listenAddress string
	api           *api.API
}

// NewServer creates a server instance.
func NewServer(listenAddress, dataDir string) (*Server, error) {
	log.Println("using data directory", dataDir)
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3",
		filepath.Join(dataDir, "crossjoin.db")+"?_txlock=immediate&cache=shared")
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	api, err := api.NewAPI(db)
	if err != nil {
		return nil, err
	}
	return &Server{
		listenAddress: listenAddress,
		api:           api,
	}, nil
}

func (s *Server) Start() error {
	log.Printf("listening on %s", s.listenAddress)
	return http.ListenAndServe(s.listenAddress, s.api.Handler())
}
