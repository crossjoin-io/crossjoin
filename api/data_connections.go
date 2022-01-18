package api

import (
	"github.com/crossjoin-io/crossjoin/config"
)

func (api *API) StoreDataConnection(connection config.DataConnection) error {
	_, err := api.db.Exec("REPLACE INTO data_connections (id, type, path, connection_string) VALUES ($1, $2, $3, $4)",
		connection.ID, connection.Type, connection.Path, connection.ConnectionString,
	)
	return err
}

func (api *API) ReadDataConnection(id string) (*config.DataConnection, error) {
	conn := &config.DataConnection{
		ID: id,
	}
	err := api.db.QueryRow("SELECT type, path, connection_string FROM data_connections WHERE id = $1", id).
		Scan(&conn.Type, &conn.Path, &conn.ConnectionString)
	if err != nil {
		return nil, err
	}
	conn.ExpandConnectionString()
	return conn, nil
}
