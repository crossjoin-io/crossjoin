package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func (api *API) postTasksResult(_ http.ResponseWriter, r *http.Request) Response {
	result := TaskResult{}
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		log.Println(err)
		return Response{
			OK:     false,
			Status: http.StatusBadRequest,
		}
	}
	if result.Output == nil {
		result.Output = map[string]interface{}{}
	}
	marshaledOutput, err := json.Marshal(result.Output)
	if err != nil {
		log.Println(err)
		return Response{
			OK:     false,
			Status: http.StatusBadRequest,
		}
	}
	_, err = api.db.Exec("update tasks set completed_at = datetime('now'), success = $1, output = $2, stdout = $3, stderr = $4 where id = $5",
		result.OK, marshaledOutput, result.Stdout, result.Stderr, result.ID)
	if err != nil {
		log.Println(err)
		return Response{
			OK:     false,
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}

	// Now we need to schedule the next task.
	workflowRunID := ""
	workflowTaskID := ""
	err = api.db.QueryRow("select workflow_run_id, workflow_task_id from tasks where id = $1", result.ID).
		Scan(&workflowRunID, &workflowTaskID)
	if err != nil {
		log.Println(err)
		return Response{
			OK:     false,
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	workflowID := ""
	err = api.db.QueryRow("select workflow_id from workflow_runs where id = $1", workflowRunID).Scan(&workflowID)
	if err != nil {
		log.Println(err)
		return Response{
			OK:     false,
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}

	// If the task wasn't a success, fail the workflow run.
	if !result.OK {
		err = api.CompleteWorkflowRun(workflowRunID, false)
		if err != nil {
			log.Println(err)
			return Response{
				OK:     false,
				Status: http.StatusInternalServerError,
				Error:  err.Error(),
			}
		}
		return Response{
			OK: true,
		}
	}

	workflow, err := api.GetWorkflow(workflowID)
	if err != nil {
		log.Println(err)
		return Response{
			OK:     false,
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	nextWorkflowTaskID := workflow.Tasks[workflowTaskID].Next
	if nextWorkflowTaskID == "" {
		// End of the workflow
		err = api.CompleteWorkflowRun(workflowRunID, true)
		if err != nil {
			log.Println(err)
			return Response{
				OK:     false,
				Status: http.StatusInternalServerError,
				Error:  err.Error(),
			}
		}
		return Response{
			OK: true,
		}
	}
	err = api.ScheduleTask(workflowRunID, nextWorkflowTaskID, result.Output)
	if err != nil {
		log.Println(err)
		return Response{
			OK:     false,
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}

	return Response{
		OK: true,
	}
}
