package api

import (
	"github.com/crossjoin-io/crossjoin/config"
)

func (api *API) StoreDataConnection(hash string, connection config.DataConnection) error {
	_, err := api.db.Exec("REPLACE INTO data_connections (config_hash, id, type, path, connection_string) VALUES ($1, $2, $3, $4, $5)",
		hash, connection.ID, connection.Type, connection.Path, connection.ConnectionString,
	)
	return err
}

func (api *API) ReadDataConnection(hash, id string) (*config.DataConnection, error) {
	conn := &config.DataConnection{
		ID: id,
	}
	err := api.db.QueryRow("SELECT type, path, connection_string FROM data_connections WHERE config_hash = $1 AND id = $2", hash, id).
		Scan(&conn.Type, &conn.Path, &conn.ConnectionString)
	if err != nil {
		return nil, err
	}
	conn.ExpandConnectionString()
	return conn, nil
}
