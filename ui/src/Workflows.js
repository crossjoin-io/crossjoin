import { useState, useEffect } from "preact/hooks";
import { html } from "htm/preact";

export function Workflows() {
  const [workflows, setWorkflows] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState();

  useEffect(() => {
    fetch("/api/workflows")
      .then((response) => {
        if (response.ok) {
          return response.json();
        }
      })
      .then((data) => {
        setLoading(false);
        setWorkflows(data.response);
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

  let workflowElems = [];
  for (id in workflows) {
    workflowElems.push(
      html`<tr>
        <td>${workflows[id].id}</td>
        <td>${Object.keys(workflows[id].tasks || {}).length} tasks</td>
        <td><a href="/app/workflows/${id}/runs">Runs</a></td>
      </tr>`
    );
  }
  return html`<h1>Workflows</h1>
    <table class="pure-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Tasks</th>
          <th>Runs</th>
        </tr>
      </thead>
      ${workflowElems}
    </table>`;
}

export function WorkflowRuns(props) {
  const [workflowRuns, setWorkflowRuns] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState();

  useEffect(() => {
    fetch(`/api/workflows/${props.workflowID}/runs`)
      .then((response) => {
        if (response.ok) {
          return response.json();
        }
      })
      .then((data) => {
        setLoading(false);
        setWorkflowRuns(data.response);
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
  workflowRuns.sort((a, b) => {
    if (a.started_at < b.started_at) {
      return 1;
    }
    return -1;
  });
  for (i in workflowRuns) {
    const run = workflowRuns[i];
    let statusIcon = run.success
      ? html`<i class="cj-green fas fa-check"></i>`
      : run.completed_at
      ? html`<i class="fas fa-times"></i>`
      : html`<i class="fas fa-hourglass"></i>`;
    runs.push(
      html`<tr>
        <td>${run.id}</td>
        <td>
          <a href="/app/workflows/${run.workflow_id}/runs/${run.id}/tasks"
            >Tasks</a
          >
        </td>
        <td>${run.started_at}</td>
        <td>${run.completed_at}</td>
        <td>${statusIcon}</td>
      </tr>`
    );
  }
  return html`<h1>Workflows</h1>
    <div class="cj-breadcrumb">
      <a href="/app/workflows">Workflows</a>
      <span> / </span>
      ${props.workflowID}
      <span> / </span>
      Runs
    </div>
    <table class="pure-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Tasks</th>
          <th>Started</th>
          <th>Completed</th>
          <th>Status</th>
        </tr>
      </thead>
      ${runs}
    </table>`;
}

export function WorkflowRunTasks(props) {
  const [workflowTasks, setWorkflowTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState();

  useEffect(() => {
    fetch(
      `/api/workflows/${props.workflowID}/runs/${props.workflowRunID}/tasks`
    )
      .then((response) => {
        if (response.ok) {
          return response.json();
        }
      })
      .then((data) => {
        setLoading(false);
        setWorkflowTasks(data.response);
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

  let tasks = [];
  for (i in workflowTasks) {
    const task = workflowTasks[i];
    let statusIcon = task.success
      ? html`<i class="cj-green fas fa-check"></i>`
      : task.completed_at
      ? html`<i class="fas fa-times"></i>`
      : html`<i class="fas fa-hourglass"></i>`;
    tasks.push(
      html`<tr>
        <td>${task.workflow_task_id}</td>
        <td>${task.started_at}</td>
        <td>${task.completed_at}</td>
        <td>
          <pre style="width: 15em; max-height: 15em; overflow: scroll;">
${JSON.stringify(task.output, 0, 2)}</pre
          >
        </td>
        <td>
          <pre style="width: 15em; overflow: scroll;">${task.stdout}</pre>
        </td>
        <td>
          <pre style="width: 15em; overflow: scroll;">${task.stderr}</pre>
        </td>
        <td>${statusIcon}</td>
      </tr>`
    );
  }
  return html`<h1>Workflows</h1>
    <div class="cj-breadcrumb">
      <a href="/app/workflows">Workflows</a>
      <span> / </span>
      <a href="/app/workflows/${props.workflowID}/runs">${props.workflowID}</a>
      <span> / </span>
      Tasks
    </div>
    <table class="pure-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Started</th>
          <th>Completed</th>
          <th>Output</th>
          <th>Stdout</th>
          <th>Stderr</th>
          <th>Status</th>
        </tr>
      </thead>
      ${tasks}
    </table>`;
}
