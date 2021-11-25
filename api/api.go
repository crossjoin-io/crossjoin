package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/crossjoin-io/crossjoin/config"
	"github.com/crossjoin-io/crossjoin/ui/public"
	"github.com/gorilla/mux"
)

type API struct {
	db     *sql.DB
	router *mux.Router
}

// NewAPI returns a new API instance.
func NewAPI(db *sql.DB, conf *config.Config) (*API, error) {
	r := mux.NewRouter()

	err := setupDatabase(db)
	if err != nil {
		return nil, err
	}

	api := &API{
		db:     db,
		router: r,
	}

	// Load the initial config
	for _, workflow := range conf.Workflows {
		err = api.StoreWorkflow(workflow)
		if err != nil {
			return nil, err
		}
	}

	return api, nil
}

func (api *API) Handler() http.Handler {
	baseMux := http.NewServeMux()
	baseMux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/app/") {
			r.URL.Path = "/"
		}
		http.FileServer(http.FS(public.Content)).ServeHTTP(w, r)
	}))
	baseMux.Handle("/ui.go", http.NotFoundHandler())
	baseMux.Handle("/api/", api.router)
	api.handle("GET", "/api/ping", func(r *http.Request) Response {
		return Response{}
	})
	api.handle("GET", "/api/db/schema", api.getDBSchema)
	api.handle("GET", "/api/tasks/poll", api.getTasksPoll)
	api.handle("POST", "/api/tasks/result", api.postTasksResult)

	api.handle("GET", "/api/workflows", api.getWorkflows)
	api.handle("GET", "/api/workflows/{workflow_id}/runs", api.getWorkflowRuns)
	api.handle("GET", "/api/workflows/{workflow_id}/runs/{workflow_run_id}/tasks", api.getWorkflowRunTasks)
	api.handle("POST", "/api/workflows/{workflow_id}/start", api.postWorkflowsStart)
	return baseMux
}

func (api *API) handle(method, route string, handler func(r *http.Request) Response) {
	api.router.Methods(method).Path(route).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("handling", r.Method, r.URL.String())
		resp := handler(r)
		if resp.Error == "" {
			resp.OK = true
		}
		if resp.Status > 0 {
			w.WriteHeader(resp.Status)
		}
		w.Header().Add("content-type", "application/json")
		enc := json.NewEncoder(w)
		enc.Encode(resp)
	})
}
