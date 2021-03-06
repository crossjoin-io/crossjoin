package api

import (
	"fmt"
	"log"
	"net/http"
)

func (api *API) getStatusSummary(_ http.ResponseWriter, r *http.Request) Response {
	hash, err := api.LatestConfigHash()
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}

	var (
		totalConnections    int
		totalDatasets       int
		totalWorkflows      int
		totalTasksCompleted int
	)

	err = api.db.QueryRow(`
	SELECT
		(SELECT COUNT(*) FROM data_connections WHERE config_hash = $1) AS total_connections,
		(SELECT COUNT(*) FROM datasets WHERE config_hash = $1) AS total_datasets,
		(SELECT COUNT(*) FROM workflows WHERE config_hash = $1) AS total_workflows,
		(SELECT COUNT(*) FROM tasks WHERE completed_at IS NOT NULL) AS total_tasks_completed
	`, hash).Scan(&totalConnections, &totalDatasets, &totalWorkflows, &totalTasksCompleted)
	if err != nil {
		log.Println(fmt.Errorf("query summary counts: %w", err))
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	runs := []SummaryTaskRun{}
	failures := []SummaryTaskRun{}

	recentRunsRows, err := api.db.Query(`SELECT
			id,
			(SELECT workflow_id FROM workflow_runs WHERE id = workflow_run_id) AS workflow_id,
			workflow_run_id,
			workflow_task_id,
			created_at,
			started_at,
			completed_at,
			success
		FROM tasks ORDER BY COALESCE(completed_at, started_at) DESC LIMIT 10`)
	if err != nil {
		log.Println(err)
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	defer recentRunsRows.Close()
	for recentRunsRows.Next() {
		run := SummaryTaskRun{}
		err = recentRunsRows.Scan(&run.ID, &run.WorkflowID, &run.WorkflowRunID, &run.WorkflowTaskID,
			&run.CreatedAt, &run.StartedAt, &run.CompletedAt, &run.Success)
		if err != nil {
			log.Println(err)
			return Response{
				Status: http.StatusInternalServerError,
				Error:  err.Error(),
			}
		}
		runs = append(runs, run)
	}

	recentFailuresRows, err := api.db.Query(`SELECT
			id,
			(SELECT workflow_id FROM workflow_runs WHERE id = workflow_run_id) AS workflow_id,
			workflow_run_id,
			workflow_task_id,
			created_at,
			started_at,
			completed_at,
			success
		FROM tasks WHERE success = 0 ORDER BY COALESCE(completed_at, started_at) DESC LIMIT 10`)
	if err != nil {
		log.Println(err)
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	defer recentFailuresRows.Close()
	for recentFailuresRows.Next() {
		run := SummaryTaskRun{}
		err = recentFailuresRows.Scan(&run.ID, &run.WorkflowID, &run.WorkflowRunID, &run.WorkflowTaskID,
			&run.CreatedAt, &run.StartedAt, &run.CompletedAt, &run.Success)
		if err != nil {
			log.Println(err)
			return Response{
				Status: http.StatusInternalServerError,
				Error:  err.Error(),
			}
		}
		failures = append(failures, run)
	}

	return Response{
		Response: StatusSummary{
			RecentTaskRuns:      runs,
			RecentTaskFailures:  failures,
			TotalConnections:    totalConnections,
			TotalDatasets:       totalDatasets,
			TotalWorkflows:      totalWorkflows,
			TotalTasksCompleted: totalTasksCompleted,
		},
	}
}
