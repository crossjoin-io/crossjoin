package api

import (
	"net/http"

	"github.com/crossjoin-io/crossjoin/config"
)

func (api *API) getDataConnections(_ http.ResponseWriter, r *http.Request) Response {
	rows, err := api.db.Query("SELECT name, type, path, connection_string FROM data_connections")
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	defer rows.Close()
	connections := []config.DataConnection{}
	for rows.Next() {
		connection := config.DataConnection{}
		err = rows.Scan(&connection.Name, &connection.Type, &connection.Path, &connection.ConnectionString)
		if err != nil {
			return Response{
				Status: http.StatusInternalServerError,
				Error:  err.Error(),
			}
		}
		connections = append(connections, connection)
	}
	return Response{
		Response: connections,
	}
}
