package api

import (
	"github.com/crossjoin-io/crossjoin/config"
	"github.com/google/uuid"
)

func (api *API) StoreWorkflow(workflow config.Workflow) error {
	_, err := api.db.Exec("INSERT INTO workflows (id, text) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		workflow.ID, workflow.String(),
	)
	return err
}

func (api *API) StartWorkflow(id string) error {
	// Create a workflow run
	workflowRunID, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	_, err = api.db.Exec(`INSERT INTO workflow_runs (id, workflow_id, started_at)
	VALUES ($1, $2, datetime('now'))`, workflowRunID.String(), id)
	if err != nil {
		return err
	}
	var workflowText []byte
	err = api.db.QueryRow(`select text from workflows where id = $1`, id).Scan(&workflowText)
	if err != nil {
		return err
	}
	workflow := &config.Workflow{}
	err = workflow.Parse(workflowText)
	if err != nil {
		return err
	}

	// Schedule the first task
	taskID, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	_, err = api.db.Exec(`insert into tasks (id, workflow_run_id, workflow_task_id, created_at) values
	($1, $2, $3, datetime('now'))`, taskID.String(), workflowRunID.String(), workflow.Start)
	return err
}
