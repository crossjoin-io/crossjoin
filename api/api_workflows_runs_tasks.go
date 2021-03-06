package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (api *API) getWorkflowRunTasks(_ http.ResponseWriter, r *http.Request) Response {
	vars := mux.Vars(r)
	workflowRunID := vars["workflow_run_id"]

	rows, err := api.db.Query(`SELECT
		id,
		workflow_task_id,
		input,
		output,
		created_at,
		started_at,
		timeout_at,
		completed_at,
		attempts_left,
		stdout,
		stderr,
		success
	FROM tasks WHERE workflow_run_id = $1`,
		workflowRunID)
	if err != nil {
		log.Println(err)
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	defer rows.Close()
	runs := []TaskRun{}
	for rows.Next() {
		run := TaskRun{
			WorkflowRunID: workflowRunID,
		}
		output := ""
		err = rows.Scan(&run.ID, &run.WorkflowTaskID, &run.Input, &output,
			&run.CreatedAt, &run.StartedAt, &run.TimeoutAt, &run.CompletedAt, &run.AttemptsLeft,
			&run.Stdout, &run.Stderr, &run.Success)
		if err != nil {
			log.Println(err)
			return Response{
				Status: http.StatusInternalServerError,
				Error:  err.Error(),
			}
		}
		run.Output = json.RawMessage(output)
		runs = append(runs, run)
	}
	return Response{
		Response: runs,
	}
}
