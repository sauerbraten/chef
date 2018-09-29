package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/sauerbraten/chef/internal/db"
	"github.com/sauerbraten/chef/internal/ips"
)

func (s *server) lookup() http.HandlerFunc {
	tmpl, err := template.
		New("base.tmpl"). // must be the base template (entry point) so templates are associated correctly by ParseFiles()
		Funcs(template.FuncMap{
			"timestring": func(timestamp int64) string { return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05") },
			"kidbanned":  func(ip string) bool { return s.kidban.IsBanned(net.ParseIP(ip)) },
		}).
		ParseFiles("templates/base.tmpl", "templates/results.tmpl")
	if err != nil {
		log.Fatalln(err)
	}

	return func(resp http.ResponseWriter, req *http.Request) {
		nameOrIP := req.FormValue("q")

		sorting := db.ByNameFrequency
		if req.FormValue("sorting") == db.ByLastSeen.Identifier {
			sorting = db.ByLastSeen
		}

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

		finishedLookup := s.db.Lookup(nameOrIP, sorting, directLookupForced)

		if req.FormValue("format") == "json" {
			err := json.NewEncoder(resp).Encode(finishedLookup)
			if err != nil {
				log.Println(err)
			}
		} else {
			err := tmpl.Execute(resp, finishedLookup)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
