import { useState, useEffect } from "preact/hooks";
import { html } from "htm/preact";
import { Card } from "./components/Card";

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

  let connectionCards = [];
  for (i in connections) {
    let icon = "plug";
    switch (connections[i].type) {
      case "csv":
        icon = "file-csv";
        break;
      case "postgres":
        icon = "database";
        break;
    }
    connectionCards.push(html`
      <${Card} header=${connections[i].id}>
        <div><i class="fa fa-${icon}"></i> ${connections[i].type}</div>
        <div>Path: ${connections[i].path}</div>
        <div>Connection string: ${connections[i].connection_string}</div>
      </${Card}>
    `);
  }
  return html`
    <${Card} header="Connections">
      ${connectionCards}
    </${Card}
    `;
}
