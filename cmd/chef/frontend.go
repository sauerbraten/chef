package main

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/sauerbraten/chef/internal/db"
	"github.com/sauerbraten/chef/pkg/extinfo"
	"github.com/sauerbraten/chef/pkg/ips"
)

type Frontend struct {
	chi.Router
	db *db.Database
}

func NewFrontend(db *db.Database) Frontend {
	return Frontend{
		Router: chi.NewRouter(),
		db:     db,
	}
}

func (f *Frontend) lookupSightings(resp http.ResponseWriter, req *http.Request, send func(http.ResponseWriter, db.FinishedLookup)) {
	nameOrIP := req.FormValue("q")

	sorting := db.ByNameFrequency
	if req.FormValue("sorting") == db.ByLastSeen.Identifier {
		sorting = db.ByLastSeen
	}

	last90DaysOnly := !(req.FormValue("search_old") == "true")

	findAliases := req.FormValue("direct") != "true"

	// (permanently) redirect partial IP queries
	if ips.IsPartialOrFullCIDR(nameOrIP) {
		subnet := ips.GetSubnet(nameOrIP)

		if nameOrIP != subnet.String() {
			u, _ := url.ParseRequestURI(req.RequestURI) // it's safe to assume this will not fail
			params := u.Query()
			params.Set("q", subnet.String())
			u.RawQuery = params.Encode()
			http.Redirect(resp, req, u.String(), http.StatusPermanentRedirect)
			return
		}
	}

	send(resp, f.db.Lookup(nameOrIP, sorting, last90DaysOnly, findAliases))
}

func (f *Frontend) lookupGames(resp http.ResponseWriter, req *http.Request, send func(http.ResponseWriter, db.FinishedGamesLookup)) {
	q := db.GamesLookupQuery{
		PlayerName: req.FormValue("player"),
		Map:        req.FormValue("map"),
		GameMode:   -1,
	}

	_gameMode := req.FormValue("mode")
	gameMode, err := strconv.Atoi(_gameMode)
	if err == nil {
		q.GameMode = extinfo.GameMode(gameMode)
	}

	if q.PlayerName == "" && q.Map == "" && q.GameMode == -1 {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	send(resp, f.db.LookupGame(q))
}

type GameWithStats struct {
	*db.Game
	Scores []db.Score
	Stats  []db.Stats
}

func (f *Frontend) fetchGame(resp http.ResponseWriter, req *http.Request, send func(http.ResponseWriter, GameWithStats)) {
	_gameID := chi.URLParam(req, "id")
	gameID, err := strconv.ParseInt(_gameID, 10, 64)
	if err != nil {
		resp.WriteHeader(http.StatusNotFound)
		return
	}

	game := f.db.GetGame(gameID)
	if game == nil {
		resp.WriteHeader(http.StatusNotFound)
		return
	}

	send(resp, GameWithStats{
		game,
		f.db.GetScores(gameID),
		f.db.GetGameStats(gameID),
	})
}
