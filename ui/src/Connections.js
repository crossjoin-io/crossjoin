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
        <td>${connections[i].Name}</td>
        <td>${connections[i].Type}</td>
        <td>${connections[i].Path}</td>
        <td>${connections[i].ConnectionString}</td>
      </tr>`
    );
  }
  return html`<h1>Connections</h1>
    <table class="pure-table">
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
