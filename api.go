package chef

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/sauerbraten/chef/db"
)

type API struct {
	http.Handler
	db *db.Database
}

func NewAPI(db *db.Database) *API {
	api := &API{
		db: db,
	}

	r := http.NewServeMux()
	r.HandleFunc("/lookup", api.lookup)
	r.HandleFunc("/server", api.server)

	api.Handler = withJSONRespHeader(r)

	return api
}

func (api *API) lookup(resp http.ResponseWriter, req *http.Request) {
	nameOrIP, sorting, last90DaysOnly, directLookupForced, redirected := parseLookupRequest(resp, req)
	if redirected {
		return
	}

	results := api.db.Lookup(nameOrIP, sorting, last90DaysOnly, directLookupForced)

	err := json.NewEncoder(resp).Encode(results)
	if err != nil {
		slog.Error("encode lookup results", "error", err)
	}
}

func (api *API) server(resp http.ResponseWriter, req *http.Request) {
	desc := req.FormValue("q")

	results := api.db.FindServerByDescription(desc)

	err := json.NewEncoder(resp).Encode(results)
	if err != nil {
		slog.Error("encode server list", "error", err)
	}
}

func withJSONRespHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(resp, req)
	})
}
