package api

import (
	"net/http"

	"github.com/crossjoin-io/crossjoin/config"
)

func (api *API) getWorkflows(r *http.Request) Response {
	rows, err := api.db.Query("SELECT id, text FROM workflows")
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	defer rows.Close()
	workflows := map[string]config.Workflow{}
	for rows.Next() {
		id, text := "", ""
		err = rows.Scan(&id, &text)
		if err != nil {
			return Response{
				Status: http.StatusInternalServerError,
				Error:  err.Error(),
			}
		}
		workflow := config.Workflow{}
		err = workflow.Parse([]byte(text))
		if err != nil {
			return Response{
				Status: http.StatusInternalServerError,
				Error:  err.Error(),
			}
		}
		workflows[id] = workflow
	}
	return Response{
		Response: workflows,
	}
}
