import { useState, useEffect } from "preact/hooks";
import { html } from "htm/preact";
import "./Datasets.css";
import { Card } from "./components/Card";

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
        <td>${datasets[i].id}</td>
        <td><pre>${datasets[i].Text}</pre></td>
        <td><a href="/app/datasets/${datasets[i].id}/preview">Preview</a></td>
      </tr>`
    );
  }
  return html`
  <${Card} header="Datasets">
    <table class="pure-table pure-table-horizontal">
      <thead>
        <tr>
          <th>ID</th>
          <th>Text</th>
          <th>Preview</th>
        </tr>
      </thead>
      ${datasetElems}
    </table>
  </${Card}>`;
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
  return html`<${Card} header="Datasets">
    <div class="cj-breadcrumb">
      <a href="/app/datasets">Datasets</a>
      <span> / </span>
      ${props.datasetID}
      <span> / </span>
      Preview
    </div>
    <div class="cj-dataset-preview">
      <table class="pure-table pure-table-horizontal">
        <thead>
          <tr>
            ${columns}
          </tr>
        </thead>
        ${rows}
      </table>
    </div></${Card}
  >`;
}
