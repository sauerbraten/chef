package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sauerbraten/chef/internal/db"
	"github.com/sauerbraten/chef/pkg/kidban"
)

type API struct {
	Frontend
}

func NewAPI(db *db.Database, kidban *kidban.Checker) *API {
	api := &API{
		Frontend: NewFrontend(db, kidban),
	}

	api.HandleFunc("/lookup", api.lookup)
	api.HandleFunc("/server", api.server)

	return api
}

func (api *API) lookup(resp http.ResponseWriter, req *http.Request) {
	api.Frontend.lookup(resp, req, func(resp http.ResponseWriter, results db.FinishedLookup) {
		err := json.NewEncoder(resp).Encode(results)
		if err != nil {
			log.Println(err)
		}
	})
}

func (api *API) server(resp http.ResponseWriter, req *http.Request) {
	desc := req.FormValue("q")

	results := api.db.FindServerByDescription(desc)
	err := json.NewEncoder(resp).Encode(results)
	if err != nil {
		log.Println(err)
	}
}
