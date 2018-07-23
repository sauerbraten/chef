package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/sauerbraten/chef/db"
	"github.com/sauerbraten/chef/ips"
	"github.com/sauerbraten/chef/kidban"
)

func TimestampToString(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05")
}

func IsInKidbannedNetwork(ipString string) bool {
	return kidban.IsInKidbannedNetwork(net.ParseIP(ipString))
}

func (s *server) lookup() http.HandlerFunc {
	tmpl, err := template.
		New("results.html").
		Funcs(template.FuncMap{"timestring": TimestampToString, "ipIsInKidbannedNetwork": IsInKidbannedNetwork}).
		ParseFiles("html/results.html")
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
