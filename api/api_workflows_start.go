package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (api *API) postWorkflowsStart(_ http.ResponseWriter, r *http.Request) Response {
	vars := mux.Vars(r)
	workflowID := vars["workflow_id"]

	var workflowInput map[string]interface{}
	json.NewDecoder(r.Body).Decode(&workflowInput)

	latestHash, err := api.LatestConfigHash()
	if err != nil {
		log.Println(err)
		return Response{
			OK:     false,
			Status: http.StatusInternalServerError,
		}
	}

	err = api.StartWorkflow(latestHash, workflowID, workflowInput)
	if err != nil {
		log.Println(err)
		return Response{
			OK:     false,
			Status: http.StatusInternalServerError,
		}
	}
	return Response{
		OK: true,
	}
}
