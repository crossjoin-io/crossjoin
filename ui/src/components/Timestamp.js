import { html } from "htm/preact";

export function Timestamp(props) {
  const str = new Intl.DateTimeFormat("en-US", {
    year: "numeric",
    month: "numeric",
    day: "numeric",
    hour: "numeric",
    minute: "numeric",
    second: "numeric",
    hour12: true,
    timeZone: "America/Los_Angeles",
  }).format(props.ts);
  return html`<span>${str}</span>`;
}
