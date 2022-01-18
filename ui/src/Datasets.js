import { useState, useEffect } from "preact/hooks";
import { html } from "htm/preact";

export function Datasets() {
  const [datasets, setDatasets] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState();

  useEffect(() => {
    fetch("/api/datasets")
      .then((response) => {
        if (response.ok) {
          return response.json();
        }
      })
      .then((data) => {
        setLoading(false);
        setDatasets(data.response);
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

  let datasetElems = [];
  for (i in datasets) {
    datasetElems.push(
      html`<tr>
        <td>${datasets[i].ID}</td>
        <td><pre>${datasets[i].Text}</pre></td>
        <td><a href="/app/datasets/${datasets[i].ID}/preview">Preview</a></td>
      </tr>`
    );
  }
  return html`<h1>Datasets</h1>
    <table class="pure-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Text</th>
          <th>Preview</th>
        </tr>
      </thead>
      ${datasetElems}
    </table>`;
}

export function DatasetPreview(props) {
  const [datasetPreview, setDatasetPreview] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState();

  useEffect(() => {
    fetch(`/api/datasets/${props.datasetID}/preview`)
      .then((response) => {
        if (response.ok) {
          return response.json();
        }
      })
      .then((data) => {
        setLoading(false);
        setDatasetPreview(data.response);
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

  let rows = [];
  if (datasetPreview.length === 0) {
    return;
  }
  const columns = Object.keys(datasetPreview[0]).map(
    (v) => html`<th>${v}</th>`
  );

  for (i in datasetPreview) {
    const row = Object.keys(datasetPreview[0]).map(
      (v) => html`<td>${datasetPreview[i][v]}</td>`
    );
    rows.push(
      html`<tr>
        ${row}
      </tr>`
    );
  }
  return html`<h1>Datasets</h1>
    <div class="cj-breadcrumb">
      <a href="/app/datasets">Datasets</a>
      <span> / </span>
      ${props.datasetID}
      <span> / </span>
      Preview
    </div>
    <table class="pure-table">
      <thead>
        <tr>
          ${columns}
        </tr>
      </thead>
      ${rows}
    </table>`;
}
