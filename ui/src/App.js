import { Router } from "preact-router";
import { useState, useEffect } from "preact/hooks";
import { html } from "htm/preact";
import "./App.css";

function Home() {
  return html`<h1>Home</h1>`;
}

function Workflows() {
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

  console.log(workflows);

  let workflowElems = [];
  for (id in workflows) {
    workflowElems.push(
      html`<tr>
        <td>${workflows[id].ID}</td>
        <td>${Object.keys(workflows[id].Tasks || {}).length} tasks</td>
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

function WorkflowRuns(props) {
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

  console.log(workflowRuns);

  let runs = [];
  for (i in workflowRuns) {
    const run = workflowRuns[i];
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
        <td>${run.success ? "✅" : "❌"}</td>
      </tr>`
    );
  }
  return html`<h1>Workflows</h1>
    <div class="cj-breadcrumb">
      <a href="/app/workflows">Workflows</a>
      <span> / </span>
      ${props.workflowID}
    </div>
    <table class="pure-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Tasks</th>
          <th>Started</th>
          <th>Completed</th>
          <th>Success</th>
        </tr>
      </thead>
      ${runs}
    </table>`;
}

function WorkflowRunTasks(props) {
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
    console.log(task);
    tasks.push(
      html`<tr>
        <td>${task.workflow_task_id}</td>
        <td>${task.started_at}</td>
        <td>${task.completed_at}</td>
        <td>
          <pre style="width: 15em; overflow: scroll;">${task.stdout}</pre>
        </td>
        <td>${task.success ? "✅" : "❌"}</td>
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
          <th>Stdout</th>
          <th>Success</th>
        </tr>
      </thead>
      ${tasks}
    </table>`;
}

function App() {
  return html`
    <div class="cj-container">
        <div class="cj-nav">
            <div class="pure-menu pure-menu-horizontal pure-menu-scrollable">
            <a href="/app" class="pure-menu-heading pure-menu-link">Crossjoin</a>
            <ul class="pure-menu-list">
                <li class="pure-menu-item">
                <a href="/app/workflows" class="pure-menu-link">Workflows</a>
                </li>
            </ul>
            </div>
        </div>
        <div class="cj-content">
            <${Router}>
                <${Home} path="/app/" />
                <${Workflows} path="/app/workflows" />
                <${WorkflowRuns} path="/app/workflows/:workflowID/runs" />
                <${WorkflowRunTasks} path="/app/workflows/:workflowID/runs/:workflowRunID/tasks" />
            </${Router}>
        </div>
    </div>
    `;
}

export default App;
