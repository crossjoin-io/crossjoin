import { useState, useEffect } from "preact/hooks";
import { html } from "htm/preact";

export function Connections() {
  const [connections, setConnections] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState();

  useEffect(() => {
    fetch("/api/data_connections")
      .then((response) => {
        if (response.ok) {
          return response.json();
        }
      })
      .then((data) => {
        setLoading(false);
        setConnections(data.response);
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

  let connectionElems = [];
  for (i in connections) {
    connectionElems.push(
      html`<tr>
        <td>${connections[i].name}</td>
        <td>${connections[i].type}</td>
        <td>${connections[i].path}</td>
        <td>${connections[i].connection_string}</td>
      </tr>`
    );
  }
  return html`<h1>Connections</h1>
    <table class="pure-table pure-table-horizontal">
      <thead>
        <tr>
          <th>Name</th>
          <th>Type</th>
          <th>Path</th>
          <th>Connection string</th>
        </tr>
      </thead>
      ${connectionElems}
    </table>`;
}
