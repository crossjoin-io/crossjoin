import { html } from "htm/preact";
import { Spinner } from "./Spinner";
import "./TaskRun.css";
import { Timestamp } from "./Timestamp";

export function TaskRun(props) {
  let statusIcon = props.run.success
    ? html`<i class="cj-green fas fa-check"></i>`
    : props.run.completed_at
    ? html`<i class="fas fa-times"></i>`
    : html`<${Spinner} />`;
  return html`
    <div class="cj-task-run-header">
      ${statusIcon} ${" "}
      <a
        href="/app/workflows/${props.run.workflow_id}/runs/${props.run
          .workflow_run_id}/tasks"
        >${props.run.workflow_id}/${props.run.workflow_task_id}</a
      >
    </div>
    <div class="cj-task-run-detail">
      <span class="cj-task-run-detail-label">Started at: </span>
      <${Timestamp} ts=${new Date(props.run.started_at)} />
    </div>
    <div class="cj-task-run-detail">
      <span class="cj-task-run-detail-label">Completed at: </span>
      <${Timestamp} ts=${new Date(props.run.completed_at)} />
    </div>
  `;
}
