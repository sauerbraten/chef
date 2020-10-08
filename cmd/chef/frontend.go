package main

import (
	"net"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/sauerbraten/chef/internal/db"
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

func (f *Frontend) lookup(resp http.ResponseWriter, req *http.Request, send func(http.ResponseWriter, db.FinishedLookup)) {
	nameOrIP := req.FormValue("q")

	sorting := db.ByNameFrequency
	if req.FormValue("sorting") == db.ByLastSeen.Identifier {
		sorting = db.ByLastSeen
	}

	last90DaysOnly := !(req.FormValue("search_old") == "true")

	directLookupForced := req.FormValue("direct") == "true"

	// (permanently) redirect partial IP queries
	if ips.IsPartialOrFullCIDR(nameOrIP) {
		var subnet *net.IPNet
		subnet = ips.GetSubnet(nameOrIP)

		if nameOrIP != subnet.String() {
			u, _ := url.ParseRequestURI(req.RequestURI) // it's safe to assume this will not fail
			params := u.Query()
			params.Set("q", subnet.String())
			u.RawQuery = params.Encode()
			http.Redirect(resp, req, u.String(), http.StatusPermanentRedirect)
			return
		}
	}

	send(resp, f.db.Lookup(nameOrIP, sorting, last90DaysOnly, directLookupForced))
}
