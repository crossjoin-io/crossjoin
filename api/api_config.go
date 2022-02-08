package api

import (
	"net/http"
)

func (api *API) postConfigReload(_ http.ResponseWriter, r *http.Request) Response {
	err := api.LoadConfig()
	if err != nil {
		return Response{
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	return Response{}
}
