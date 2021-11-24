package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (api *API) postWorkflowsStart(r *http.Request) Response {
	vars := mux.Vars(r)
	workflowID := vars["workflow_id"]
	err := api.StartWorkflow(workflowID)
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
