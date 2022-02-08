package api

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"path"
	"path/filepath"

	"github.com/crossjoin-io/crossjoin/config"
)

func (api *API) LoadConfig() error {
	conf := &config.Config{}

	switch api.configSource {
	case "file":
		log.Println("using config file", api.configPath)

		configFileContent, err := ioutil.ReadFile(api.configPath)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		absPath, err := filepath.Abs(api.configPath)
		if err != nil {
			return fmt.Errorf("get abs path: %w", err)
		}
		err = conf.Parse(configFileContent, filepath.Dir(absPath))
		if err != nil {
			return fmt.Errorf("parse config: %w", err)
		}
	case "github":
		parsedURL, err := url.Parse(api.configPath)
		if err != nil {
			return fmt.Errorf("parse config path: %w", err)
		}
		basePath := parsedURL
		basePath.Path = path.Dir(basePath.Path)
		log.Println("using GitHub file", api.configPath)
		configFileContent, err := api.fetchGitHubFile(api.configPath)
		if err != nil {
			return fmt.Errorf("read file from github: %w", err)
		}
		err = conf.Parse(configFileContent, basePath.String())
		if err != nil {
			return fmt.Errorf("parse config: %w", err)
		}
	}

	hash := conf.Hash()
	var x int
	err := api.db.QueryRow("SELECT 1 FROM configs WHERE hash = $1", hash).Scan(&x)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("query config: %w", err)
		}

		// Config doesn't exist.
		_, err = api.db.Exec("INSERT INTO configs (loaded_at, hash, config) VALUES (datetime('now'), $1, $2)", hash, conf.JSON())
		if err != nil {
			return fmt.Errorf("store config: %w", err)
		}

		// Load the config
		for _, workflow := range conf.Workflows {
			err := api.StoreWorkflow(hash, workflow)
			if err != nil {
				return err
			}
		}
		for _, connection := range conf.DataConnections {
			err := api.StoreDataConnection(hash, connection)
			if err != nil {
				return err
			}
		}
		for _, dataset := range conf.Datasets {
			err := api.StoreDataset(hash, dataset)
			if err != nil {
				return err
			}
		}
	}

	// Config already exists with the hash.
	_, err = api.db.Exec("UPDATE configs SET loaded_at = datetime('now') WHERE hash = $1", hash)
	return err
}

func (api *API) LatestConfigHash() (string, error) {
	hash := ""
	err := api.db.QueryRow("SELECT hash FROM configs ORDER BY loaded_at DESC LIMIT 1").Scan(&hash)
	return hash, err
}
