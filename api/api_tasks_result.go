package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func (api *API) postTasksResult(r *http.Request) Response {
	result := TaskResult{}
	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		log.Println(err)
		return Response{
			OK:     false,
			Status: http.StatusBadRequest,
		}
	}
	if result.Output == nil {
		result.Output = json.RawMessage("{}")
	}
	_, err = api.db.Exec("update tasks set completed_at = datetime('now'), success = $1, output = $2, stdout = $3, stderr = $4 where id = $5",
		result.OK, result.Output, result.Stdout, result.Stderr, result.ID)
	if err != nil {
		log.Println(err)
		return Response{
			OK:     false,
			Status: http.StatusInternalServerError,
			Error:  err.Error(),
		}
	}
	return Response{
		OK: true,
	}
}
