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
