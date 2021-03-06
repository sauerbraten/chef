package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"

	"github.com/sauerbraten/chef/internal/db"
)

type API struct {
	Frontend
}

func NewAPI(db *db.Database) *API {
	api := &API{
		Frontend: NewFrontend(db),
	}

	api.Use(middleware.SetHeader("Content-Type", "application/json; charset=utf-8"))

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
