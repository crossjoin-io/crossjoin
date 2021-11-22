import { render, Component } from "preact";
import { html } from "htm/preact";

class App extends Component {
  constructor() {
    super();

    this.state = { loading: true, error: null, apiOK: false };
  }

  componentWillMount() {
    let setState = this.setState.bind(this);
    fetch("/api/ping")
      .then((response) => {
        if (response.ok) {
          return response.json();
        }
      })
      .then((data) => {
        setState({
          loading: false,
          apiOK: data.ok,
        });
      })
      .catch((e) => {
        setState({
          loading: false,
          error: e.toString(),
        });
      });
  }

  render() {
    if (this.state.loading) {
      return html`Loading...`;
    }
    if (this.state.error) {
      return html`API error: ${this.state.error}`;
    }
    return html`<p>API OK.</p>`;
  }
}

render(html`<${App} />`, document.getElementById("app"));
