import { Router } from "preact-router";
import { html } from "htm/preact";
import { Connections } from "./Connections";
import { Datasets, DatasetPreview } from "./Datasets";
import { Workflows, WorkflowRuns, WorkflowRunTasks } from "./Workflows";
import "./App.css";

function Home() {
  return html`<h1>Home</h1>`;
}

function App() {
  return html`
    <div class="cj-container">
      <div class="cj-nav">
        <div class="pure-menu pure-menu-horizontal pure-menu-scrollable">
          <a href="/app" class="pure-menu-heading pure-menu-link">Crossjoin</a>
          <ul class="pure-menu-list">
            <li class="pure-menu-item">
              <a href="/app/connections" class="pure-menu-link">Connections</a>
            </li>
            <li class="pure-menu-item">
              <a href="/app/datasets" class="pure-menu-link">Datasets</a>
            </li>
            <li class="pure-menu-item">
              <a href="/app/workflows" class="pure-menu-link">Workflows</a>
            </li>
          </ul>
        </div>
      </div>
      <div class="cj-content">
        <${Router}>
          <${Home} path="/app/" />
          <${Connections} path="/app/connections" />
          <${Datasets} path="/app/datasets" />
          <${DatasetPreview} path="/app/datasets/:datasetID/preview" />
          <${Workflows} path="/app/workflows" />
          <${WorkflowRuns} path="/app/workflows/:workflowID/runs" />
          <${WorkflowRunTasks} path="/app/workflows/:workflowID/runs/:workflowRunID/tasks" />
        </${Router}>
      </div>
    </div>
  `;
}

export default App;
