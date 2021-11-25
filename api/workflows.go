package api

import (
	"encoding/json"

	"github.com/crossjoin-io/crossjoin/config"
	"github.com/google/uuid"
)

func (api *API) StoreWorkflow(workflow config.Workflow) error {
	_, err := api.db.Exec("INSERT INTO workflows (id, text) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		workflow.ID, workflow.String(),
	)
	return err
}

func (api *API) StartWorkflow(id string, workflowInput map[string]interface{}) error {
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

	workflow, err := api.GetWorkflow(id)
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

func (api *API) GetWorkflow(id string) (*config.Workflow, error) {
	var workflowText []byte
	err := api.db.QueryRow(`select text from workflows where id = $1`, id).Scan(&workflowText)
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

func (api *API) GetWorkflowFromWorkflowRunID(workflowRunID string) (*config.Workflow, error) {
	var workflowText []byte
	var workflowID string
	err := api.db.QueryRow(`select workflows.id, text from workflows join workflow_runs on workflow_runs.workflow_id = workflows.id where workflow_runs.id = $1`,
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
