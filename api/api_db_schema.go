package api

import "net/http"

func (api *API) getDBSchema(r *http.Request) Response {
	rows, err := api.db.Query("SELECT name, sql FROM sqlite_master WHERE type = $1", "table")
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	defer rows.Close()
	tables := map[string]string{}
	for rows.Next() {
		name, sql := "", ""
		err = rows.Scan(&name, &sql)
		if err != nil {
			return Response{
				Status: http.StatusInternalServerError,
				Error:  err.Error(),
			}
		}
		tables[name] = sql
	}
	return Response{
		Response: tables,
	}
}
