package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (api *API) getWorkflowRuns(_ http.ResponseWriter, r *http.Request) Response {
	vars := mux.Vars(r)
	workflowID := vars["workflow_id"]

	rows, err := api.db.Query("SELECT id, started_at, completed_at, success FROM workflow_runs WHERE workflow_id = $1",
		workflowID)
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	defer rows.Close()
	runs := []WorkflowRun{}
	for rows.Next() {
		run := WorkflowRun{
			WorkflowID: workflowID,
		}
		err = rows.Scan(&run.ID, &run.StartedAt, &run.CompletedAt, &run.Success)
		if err != nil {
			return Response{
				Status: http.StatusInternalServerError,
				Error:  err.Error(),
			}
		}
		runs = append(runs, run)
	}
	return Response{
		Response: runs,
	}
}
