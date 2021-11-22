package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/crossjoin-io/crossjoin/ui/public"
	"github.com/gorilla/mux"
)

type API struct {
	db     *sql.DB
	router *mux.Router
}

// NewAPI returns a new API instance.
func NewAPI(db *sql.DB) (*API, error) {
	r := mux.NewRouter()

	err := setupDatabase(db)
	if err != nil {
		return nil, err
	}

	return &API{
		db:     db,
		router: r,
	}, nil
}

func (api *API) Handler() http.Handler {
	baseMux := http.NewServeMux()
	baseMux.Handle("/", http.FileServer(http.FS(public.Content)))
	baseMux.Handle("/ui.go", http.NotFoundHandler())
	baseMux.Handle("/api/", api.router)
	api.handle("GET", "/api/ping", func(r *http.Request) Response {
		return Response{}
	})
	api.handle("GET", "/api/db/schema", api.getDBSchema)
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
		enc := json.NewEncoder(w)
		enc.Encode(resp)
	})
}
