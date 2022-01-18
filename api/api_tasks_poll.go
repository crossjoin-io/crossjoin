package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/crossjoin-io/crossjoin/config"
)

func (api *API) getTasksPoll(_ http.ResponseWriter, r *http.Request) Response {
	tx, err := api.db.Begin()
	if err != nil {
		if strings.Contains(err.Error(), "database is locked") {
			log.Println(err)
			return api.getTasksPoll(nil, r)
		}
		log.Println(err)
		return Response{
			OK:     false,
			Error:  err.Error(),
			Status: http.StatusInternalServerError,
		}
	}
	var t Task
	workflowRunID := ""
	workflowTaskID := ""
	var taskInput []byte
	log.Println("querying for a task")
	err = api.db.QueryRow("select id, workflow_run_id, workflow_task_id, input from tasks where "+
		"completed_at is null and "+
		"attempts_left > 0 and "+
		"(started_at is null OR timeout_at < datetime('now')) "+
		"limit 1").
		Scan(&t.ID, &workflowRunID, &workflowTaskID, &taskInput)
	if err == sql.ErrNoRows {
		log.Println("no tasks; returning")
		_ = tx.Commit()
		time.Sleep(time.Second)
		return Response{
			OK: true,
		}
	}
	if err != nil {
		log.Println(err)
		_ = tx.Commit()
		return Response{
			OK:    false,
			Error: err.Error(),
		}
	}
	err = json.Unmarshal(taskInput, &t.Input)
	if err != nil {
		log.Println(err)
		_ = tx.Commit()
		return Response{
			OK:    false,
			Error: err.Error(),
		}
	}

	// Get the workflow for the task
	var workflowText []byte
	err = api.db.QueryRow(`
	select text from workflows
	join workflow_runs on workflow_runs.workflow_id = workflows.id
	and workflow_runs.id = $1`, workflowRunID,
	).Scan(&workflowText)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return Response{
			OK:     false,
			Error:  err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	workflow := &config.Workflow{}
	workflow.Parse(workflowText)
	task := workflow.Tasks[workflowTaskID]
	t.Image = task.Image
	t.Script = task.Script
	t.Env = task.Env
	t.Datasets = task.WithDatasets

	log.Println("marking task as started")
	_, err = tx.Exec("update tasks set started_at = datetime('now'), "+
		"timeout_at = datetime('now', '+5 minute'), "+
		"attempts_left = attempts_left-1 "+
		"where id = $1", t.ID)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return Response{
			OK:     false,
			Error:  err.Error(),
			Status: http.StatusInternalServerError,
		}
	}
	tx.Commit()
	return Response{
		OK:       true,
		Response: t,
	}
}
