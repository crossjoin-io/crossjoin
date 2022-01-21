package api

import (
	"encoding/json"
	"time"
)

type Task struct {
	ID       string                 `json:"id"`
	Image    string                 `json:"image"`
	Script   string                 `json:"script"`
	Env      map[string]string      `yaml:"env"`
	Datasets []string               `json:"datasets"`
	Input    map[string]interface{} `yaml:"input"`
}

type TaskResult struct {
	ID     string                 `json:"id"`
	OK     bool                   `json:"ok"`
	Output map[string]interface{} `json:"output"`
	Stdout string                 `json:"stdout"`
	Stderr string                 `json:"stderr"`
}

type WorkflowRun struct {
	ID          string     `json:"id"`
	WorkflowID  string     `json:"workflow_id"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Success     *bool      `json:"success"`
}

type TaskRun struct {
	ID             string          `json:"id"`
	WorkflowRunID  string          `json:"workflow_run_id"`
	WorkflowTaskID string          `json:"workflow_task_id"`
	Input          json.RawMessage `json:"input"`
	Output         json.RawMessage `json:"output"`
	CreatedAt      time.Time       `json:"created_at"`
	StartedAt      *time.Time      `json:"started_at"`
	TimeoutAt      *time.Time      `json:"timeout_at"`
	CompletedAt    *time.Time      `json:"completed_at"`
	AttemptsLeft   int             `json:"attempts_left"`
	Stdout         *string         `json:"stdout"`
	Stderr         *string         `json:"stderr"`
	Success        *bool           `json:"success"`
}

type DataConnection struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Path             string `json:"path"`
	ConnectionString string `json:"connection_string"`
}

type StatusSummary struct {
	RecentTaskRuns      []SummaryTaskRun `json:"recent_task_runs"`
	RecentTaskFailures  []SummaryTaskRun `json:"recent_task_failures"`
	TotalConnections    int              `json:"total_connections"`
	TotalDatasets       int              `json:"total_datasets"`
	TotalWorkflows      int              `json:"total_workflows"`
	TotalTasksCompleted int              `json:"total_tasks_completed"`
}

type SummaryTaskRun struct {
	ID             string     `json:"id"`
	WorkflowID     string     `json:"workflow_id"`
	WorkflowRunID  string     `json:"workflow_run_id"`
	WorkflowTaskID string     `json:"workflow_task_id"`
	CreatedAt      time.Time  `json:"created_at"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	Success        *bool      `json:"success"`
}

type Response struct {
	OK       bool        `json:"ok"`
	Status   int         `json:"-"`
	Error    string      `json:"error,omitempty"`
	Response interface{} `json:"response,omitempty"`

	// internal
	customResponse bool
}

func CustomResponse() Response {
	return Response{
		customResponse: true,
	}
}
