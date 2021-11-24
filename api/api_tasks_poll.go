package api

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/crossjoin-io/crossjoin/config"
)

func (api *API) getTasksPoll(r *http.Request) Response {
	tx, err := api.db.Begin()
	if err != nil {
		if strings.Contains(err.Error(), "database is locked") {
			log.Println(err)
			return api.getTasksPoll(r)
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
	log.Println("querying for a task")
	err = api.db.QueryRow("select id, workflow_run_id, workflow_task_id, input from tasks where "+
		"completed_at is null and "+
		"attempts_left > 0 and "+
		"(started_at is null OR timeout_at < datetime('now')) "+
		"limit 1").
		Scan(&t.ID, &workflowRunID, &workflowTaskID, &t.Input)
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
