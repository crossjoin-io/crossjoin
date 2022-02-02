import { html } from "htm/preact";
import "./Card.css";

export function Card(props) {
  return html`<div class="cj-card">
    ${props.header
      ? html`<div class="cj-card-header">${props.header}</div>`
      : ""}
    ${props.children}
  </div>`;
}
