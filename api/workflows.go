package api

import (
	"encoding/json"
	"fmt"

	"github.com/crossjoin-io/crossjoin/config"
	"github.com/google/uuid"
)

func (api *API) StoreWorkflow(hash string, workflow config.Workflow) error {
	_, err := api.db.Exec("REPLACE INTO workflows (config_hash, id, text) VALUES ($1, $2, $3)",
		hash, workflow.ID, workflow.String(),
	)
	return err
}

func (api *API) StartWorkflow(hash, id string, workflowInput map[string]interface{}) error {
	// Create a workflow run
	workflowRunID, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	_, err = api.db.Exec(`INSERT INTO workflow_runs (id, config_hash, workflow_id, started_at)
	VALUES ($1, $2, $3, datetime('now'))`, workflowRunID.String(), hash, id)
	if err != nil {
		return err
	}

	workflow, err := api.GetWorkflow(hash, id)
	if err != nil {
		return err
	}

	return api.ScheduleTask(workflowRunID.String(), workflow.Start, workflowInput)
}

func (api *API) CompleteWorkflowRun(id string, success bool) error {
	_, err := api.db.Exec("UPDATE workflow_runs SET completed_at = datetime('now'), success = $1 WHERE id = $2",
		success, id)
	return err
}

func (api *API) GetWorkflow(hash, id string) (*config.Workflow, error) {
	var workflowText []byte
	err := api.db.QueryRow(`select text from workflows where config_hash = $1 AND id = $2`, hash, id).Scan(&workflowText)
	if err != nil {
		return nil, err
	}
	workflow := &config.Workflow{}
	err = workflow.Parse(workflowText)
	if err != nil {
		return nil, err
	}
	return workflow, nil
}

func (api *API) GetWorkflows(hash string) (map[string]config.Workflow, error) {
	rows, err := api.db.Query("SELECT text FROM workflows WHERE config_hash = $1", hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	workflows := map[string]config.Workflow{}
	for rows.Next() {
		var workflowText []byte
		err = rows.Scan(&workflowText)
		if err != nil {
			return nil, err
		}
		workflow := config.Workflow{}
		err = workflow.Parse(workflowText)
		if err != nil {
			return nil, fmt.Errorf("parse workflow: %w", err)
		}
		workflows[workflow.ID] = workflow
	}
	return workflows, nil
}

func (api *API) GetWorkflowFromWorkflowRunID(workflowRunID string) (*config.Workflow, error) {
	var workflowText []byte
	var workflowID string
	err := api.db.QueryRow(`select workflows.id, text from workflows join workflow_runs on (workflow_runs.config_hash, workflow_runs.workflow_id) = (workflows.config_hash, workflows.id) where workflow_runs.id = $1`,
		workflowRunID).
		Scan(&workflowID, &workflowText)
	if err != nil {
		return nil, err
	}
	workflow := &config.Workflow{}
	err = workflow.Parse(workflowText)
	if err != nil {
		return nil, err
	}
	return workflow, nil
}

func (api *API) ScheduleTask(workflowRunID string, workflowTaskID string, input map[string]interface{}) error {
	taskID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	workflow, err := api.GetWorkflowFromWorkflowRunID(workflowRunID)
	if err != nil {
		return err
	}
	taskDef := workflow.Tasks[workflowTaskID]
	taskInput := taskDef.With
	if taskInput == nil {
		taskInput = map[string]interface{}{}
	}
	for k, v := range input {
		taskInput[k] = v
	}

	marshaledTaskInput, err := json.Marshal(taskInput)
	if err != nil {
		return err
	}
	_, err = api.db.Exec(`insert into tasks (id, workflow_run_id, workflow_task_id, input, created_at) values
	($1, $2, $3, $4, datetime('now'))`, taskID.String(), workflowRunID, workflowTaskID, marshaledTaskInput)
	return err
}
