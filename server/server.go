package server

import (
	"database/sql"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/crossjoin-io/crossjoin/api"
	"github.com/crossjoin-io/crossjoin/config"
	"github.com/crossjoin-io/crossjoin/runner"
	_ "github.com/mattn/go-sqlite3"
)

// Server is an API server.
type Server struct {
	listenAddress string
	api           *api.API
	runner        bool
}

// NewServer creates a server instance.
func NewServer(listenAddress, dataDir, configFile string, runner bool) (*Server, error) {
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

	conf := &config.Config{}
	if configFile != "" {
		log.Println("using config file", configFile)

		configFileContent, err := ioutil.ReadFile(configFile)
		if err != nil {
			return nil, err
		}
		err = conf.Parse(configFileContent)
		if err != nil {
			return nil, err
		}
	}

	api, err := api.NewAPI(db, conf)
	if err != nil {
		return nil, err
	}
	return &Server{
		listenAddress: listenAddress,
		api:           api,
		runner:        runner,
	}, nil
}

func (s *Server) Start() error {
	log.Printf("listening on %s", s.listenAddress)
	if s.runner {
		go func() {
			time.Sleep(2 * time.Second)
			apiURL := "http://" + s.listenAddress
			log.Printf("starting runner polling %s", apiURL)
			workflowRunner, err := runner.NewRunner(apiURL)
			if err != nil {
				log.Fatal(err)
			}
			err = workflowRunner.Start()
			if err != nil {
				log.Fatalf("runner stopped: %v", err)
			}
		}()
	}
	return http.ListenAndServe(s.listenAddress, s.api.Handler())
}
