package api

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func (api *API) getDatasetDownload(w http.ResponseWriter, r *http.Request) Response {
	vars := mux.Vars(r)
	datasetName := vars["dataset_name"]

	filename := filepath.Join(api.dataDir, datasetName+".db")
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return Response{
				Status: http.StatusNotFound,
			}
		}
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	defer f.Close()

	w.Header().Add("content-type", "application/vnd.sqlite3")
	_, err = io.Copy(w, f)
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}

	return CustomResponse()
}
