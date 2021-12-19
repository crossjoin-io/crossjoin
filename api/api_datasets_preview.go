package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (api *API) getDatasetPreview(r *http.Request) Response {
	vars := mux.Vars(r)
	datasetName := vars["dataset_name"]

	result, err := api.PreviewDataset(datasetName)
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}

	return Response{
		Response: result,
	}
}
