import { render } from "preact";
import { html } from "htm/preact";
import App from "./App";

render(html`<${App} />`, document.getElementById("app"));
