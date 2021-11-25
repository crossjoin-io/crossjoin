import { render } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Router, route } from "preact-router";
import { html } from "htm/preact";

function Home() {
  return html` <p>Home</p>
    <a href="/app/workflows">Workflows</a>`;
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
      html`<div>
        ${workflows[id].ID} - <a href="/app/workflows/${id}/runs">Runs</a>
        <div>${JSON.stringify(workflows[id])}</div>
      </div>`
    );
  }
  return html`<div>${workflowElems}</div>`;
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
      html`<div>
        <span>${run.id} - </span>
        <a href="/app/workflows/${run.workflow_id}/runs/${run.id}/tasks"
          >Tasks</a
        >
      </div>`
    );
  }
  return html`<div>${runs}</div>`;
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
      html`<div>
        ${task.id} - ${task.workflow_task_id} - ${JSON.stringify(task.input)} -
        ${JSON.stringify(task.output)}
      </div>`
    );
  }
  return html`<div>${tasks}</div>`;
}

function App() {
  return html`
  <${Router}>
    <${Home} path="/app/" />
    <${Workflows} path="/app/workflows" />
    <${WorkflowRuns} path="/app/workflows/:workflowID/runs" />
    <${WorkflowRunTasks} path="/app/workflows/:workflowID/runs/:workflowRunID/tasks" />
  </${Router}>
  `;
}

render(html`<${App} />`, document.getElementById("app"));
