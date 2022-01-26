package api

import (
	"net/http"

	"github.com/crossjoin-io/crossjoin/config"
	"github.com/gorilla/mux"
)

func (api *API) getWorkflows(_ http.ResponseWriter, r *http.Request) Response {
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

func (api *API) getWorkflow(_ http.ResponseWriter, r *http.Request) Response {
	vars := mux.Vars(r)
	workflowID := vars["workflow_id"]

	text := ""
	err := api.db.QueryRow("SELECT text FROM workflows WHERE id = $1", workflowID).Scan(&text)
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}

	workflow := config.Workflow{
		ID: workflowID,
	}
	err = workflow.Parse([]byte(text))
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	return Response{
		Response: workflow,
	}
}
