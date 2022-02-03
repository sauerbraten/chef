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

	api.HandleFunc("/sightings", api.sightings)
	api.HandleFunc("/servers", api.servers)
	api.HandleFunc("/games", api.games)
	api.HandleFunc("/games/{id}", api.game)

	return api
}

func (api *API) sightings(resp http.ResponseWriter, req *http.Request) {
	api.Frontend.lookupSightings(resp, req, func(resp http.ResponseWriter, result db.FinishedLookup) {
		err := json.NewEncoder(resp).Encode(result)
		if err != nil {
			log.Println(err)
		}
	})
}

func (api *API) servers(resp http.ResponseWriter, req *http.Request) {
	desc := req.FormValue("desc")

	result := api.db.FindServerByDescription(desc)

	err := json.NewEncoder(resp).Encode(result)
	if err != nil {
		log.Println(err)
	}
}

func (api *API) games(resp http.ResponseWriter, req *http.Request) {
	api.Frontend.lookupGames(resp, req, func(resp http.ResponseWriter, result db.FinishedGamesLookup) {
		err := json.NewEncoder(resp).Encode(result)
		if err != nil {
			log.Println(err)
		}
	})
}

func (api *API) game(resp http.ResponseWriter, req *http.Request) {
	api.Frontend.fetchGame(resp, req, func(resp http.ResponseWriter, result GameWithStats) {
		err := json.NewEncoder(resp).Encode(result)
		if err != nil {
			log.Println(err)
		}
	})
}
