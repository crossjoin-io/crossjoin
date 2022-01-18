package api

import (
	"net/http"

	"github.com/crossjoin-io/crossjoin/config"
	"gopkg.in/yaml.v2"
)

func (api *API) getDatasets(_ http.ResponseWriter, r *http.Request) Response {
	rows, err := api.db.Query("SELECT id, text FROM datasets")
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	defer rows.Close()

	type datasetResponse struct {
		Text string
		config.Dataset
	}
	datasets := []datasetResponse{}
	for rows.Next() {
		dataset := datasetResponse{}
		err = rows.Scan(&dataset.ID, &dataset.Text)
		if err != nil {
			return Response{
				Status: http.StatusInternalServerError,
				Error:  err.Error(),
			}
		}
		err = yaml.Unmarshal([]byte(dataset.Text), &dataset.Dataset)
		if err != nil {
			return Response{
				Status: http.StatusInternalServerError,
				Error:  err.Error(),
			}
		}
		datasets = append(datasets, dataset)
	}
	return Response{
		Response: datasets,
	}
}
