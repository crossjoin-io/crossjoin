package api

import (
	"net/http"
)

func (api *API) getDataConnections(_ http.ResponseWriter, r *http.Request) Response {
	hash, err := api.LatestConfigHash()
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	rows, err := api.db.Query("SELECT id, type, path, connection_string FROM data_connections WHERE config_hash = $1", hash)
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	defer rows.Close()
	connections := []DataConnection{}
	for rows.Next() {
		connection := DataConnection{}
		err = rows.Scan(&connection.ID, &connection.Type, &connection.Path, &connection.ConnectionString)
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
