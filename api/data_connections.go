package api

import (
	"github.com/crossjoin-io/crossjoin/config"
)

func (api *API) StoreDataConnection(connection config.DataConnection) error {
	_, err := api.db.Exec("REPLACE INTO data_connections (name, type, path, connection_string) VALUES ($1, $2, $3, $4)",
		connection.Name, connection.Type, connection.Path, connection.ConnectionString,
	)
	return err
}

func (api *API) ReadDataConnection(name string) (*config.DataConnection, error) {
	conn := &config.DataConnection{
		Name: name,
	}
	err := api.db.QueryRow("SELECT type, path, connection_string FROM data_connections WHERE name = $1", name).
		Scan(&conn.Type, &conn.Path, &conn.ConnectionString)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
