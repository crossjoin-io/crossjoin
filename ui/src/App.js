import { Router } from "preact-router";
import { html } from "htm/preact";
import { useState, useEffect } from "preact/hooks";
import { Connections } from "./Connections";
import { Datasets, DatasetPreview } from "./Datasets";
import {
  Workflows,
  WorkflowDetails,
  WorkflowRuns,
  WorkflowRunTasks,
} from "./Workflows";
import { Spinner } from "./components/Spinner";
import "./App.css";
import "./Dashboard.css";
import { Card } from "./components/Card";

function Dashboard() {
  const [summary, setSummary] = useState({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState();

  useEffect(() => {
    fetch("/api/status/summary")
      .then((response) => {
        if (response.ok) {
          return response.json();
        }
      })
      .then((data) => {
        setLoading(false);
        setSummary(data.response);
      })
      .catch((e) => {
        setLoading(false);
        setError(e.toString());
      });
  }, []);

  if (loading) {
    return html`Loading...`;
  }
  if (error) {
    return html`Error: ${error}`;
  }

  let runs = [];
  for (i in summary.recent_task_runs) {
    const run = summary.recent_task_runs[i];
    let statusIcon = run.success
      ? html`<i class="cj-green fas fa-check"></i>`
      : run.completed_at
      ? html`<i class="fas fa-times"></i>`
      : html`<${Spinner} />`;
    runs.push(
      html`<tr>
        <td>${run.workflow_id}</td>
        <td>
          <a
            href="/app/workflows/${run.workflow_id}/runs/${run.workflow_run_id}/tasks"
            >${run.workflow_task_id}</a
          >
        </td>
        <td>${run.started_at}</td>
        <td>${run.completed_at}</td>
        <td>${statusIcon}</td>
      </tr>`
    );
  }

  let failures = [];
  for (i in summary.recent_task_failures) {
    const run = summary.recent_task_failures[i];
    let statusIcon = run.success
      ? html`<i class="cj-green fas fa-check"></i>`
      : run.completed_at
      ? html`<i class="fas fa-times"></i>`
      : html`<${Spinner} />`;
    failures.push(
      html`<tr>
        <td>${run.workflow_id}</td>
        <td>
          <a
            href="/app/workflows/${run.workflow_id}/runs/${run.workflow_run_id}/tasks"
            >${run.workflow_task_id}</a
          >
        </td>
        <td>${run.started_at}</td>
        <td>${run.completed_at}</td>
        <td>${statusIcon}</td>
      </tr>`
    );
  }
  return html`
    <${Card} header="Summary">
    <div class="pure-g cj-summary-counts">
      <div class="pure-u-1-4 cj-summary-count-box">
        <div class="cj-summary-count-stat">${summary.total_connections}</div>
        ${" "} connection${summary.total_connections > 1 ? "s" : ""}
      </div>
      <div class="pure-u-1-4 cj-summary-count-box">
        <div class="cj-summary-count-stat">${summary.total_datasets}</div>
        ${" "} dataset${summary.total_datasets > 1 ? "s" : ""}
      </div>
      <div class="pure-u-1-4 cj-summary-count-box">
        <div class="cj-summary-count-stat">${summary.total_workflows}</div>
        ${" "} workflow${summary.total_workflows > 1 ? "s" : ""}
      </div>
      <div class="pure-u-1-4 cj-summary-count-box">
        <div class="cj-summary-count-stat">
          ${summary.total_tasks_completed}
        </div>
        ${" "} task${summary.total_tasks_completed > 1 ? "s" : ""} completed
      </div>
    </div>
    </${Card}>

    <${Card} header="Recent tasks">
      <table class="pure-table pure-table-horizontal">
        <thead>
          <tr>
            <th>Workflow</th>
            <th>Task</th>
            <th>Started</th>
            <th>Completed</th>
            <th>Status</th>
          </tr>
        </thead>
        ${runs}
      </table>
    </${Card}>

    <${Card} header="Recent failures">
    <table class="pure-table pure-table-horizontal">
      <thead>
        <tr>
          <th>Workflow</th>
          <th>Task</th>
          <th>Started</th>
          <th>Completed</th>
          <th>Status</th>
        </tr>
      </thead>
      ${failures}
    </table>
    </${Card}>`;
}

function Nav() {
  return html`
    <div class="cj-nav">
      <div class="pure-menu pure-menu-horizontal pure-menu-scrollable">
        <a href="/app" class="pure-menu-heading pure-menu-link"
          ><i class="fas fa-xmark"></i> Crossjoin</a
        >
        <ul class="pure-menu-list">
          <li class="pure-menu-item">
            <a href="/app/connections" class="pure-menu-link"
              ><i class="fas fa-plug"></i> Connections</a
            >
          </li>
          <li class="pure-menu-item">
            <a href="/app/datasets" class="pure-menu-link"
              ><i class="fas fa-table"></i> Datasets</a
            >
          </li>
          <li class="pure-menu-item">
            <a href="/app/workflows" class="pure-menu-link"
              ><i class="fas fa-diagram-project"></i> Workflows</a
            >
          </li>
        </ul>
      </div>
    </div>
  `;
}

function App() {
  return html`
    <${Nav} />
    <div class="cj-container">
      <div class="cj-content">
        <${Router}>
          <${Dashboard} path="/app/" />
          <${Connections} path="/app/connections" />
          <${Datasets} path="/app/datasets" />
          <${DatasetPreview} path="/app/datasets/:datasetID/preview" />
          <${Workflows} path="/app/workflows" />
          <${WorkflowDetails} path="/app/workflows/:workflowID" />
          <${WorkflowRuns} path="/app/workflows/:workflowID/runs" />
          <${WorkflowRunTasks} path="/app/workflows/:workflowID/runs/:workflowRunID/tasks" />
        </${Router}>
      </div>
    </div>
  `;
}

export default App;
