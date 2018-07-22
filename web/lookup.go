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

func lookup(resp http.ResponseWriter, req *http.Request) {
	logRequest(req)

	nameOrIP := req.FormValue("q")

	sorting := db.ByNameFrequency
	if req.FormValue("sorting") == "last_seen" {
		sorting = db.ByLastSeen
	}

	directLookupForced := false
	if req.FormValue("direct") == "true" {
		directLookupForced = true
	}

	// (permanently) redirect partial IP queries
	if ips.IsPartialOrFullCIDR(nameOrIP) {
		var subnet *net.IPNet
		subnet = ips.GetSubnet(nameOrIP)

		if nameOrIP != subnet.String() {
			u, _ := url.ParseRequestURI(req.RequestURI) // safe to assume this will not fail
			params := u.Query()
			params.Set("q", subnet.String())
			u.RawQuery = params.Encode()
			http.Redirect(resp, req, u.String(), http.StatusPermanentRedirect)
			return
		}
	}

	finishedLookup := storage.Lookup(nameOrIP, sorting, directLookupForced)

	if req.FormValue("format") == "json" {
		enc := json.NewEncoder(resp)
		err := enc.Encode(finishedLookup.Results)
		if err != nil {
			log.Println(err)
		}
	} else {
		resultsTempl := template.New("results.html")
		resultsTempl = resultsTempl.Funcs(template.FuncMap{"timestring": TimestampToString, "ipIsInKidbannedNetwork": IsInKidbannedNetwork})
		resultsTempl = template.Must(resultsTempl.ParseFiles("html/results.html"))
		err := resultsTempl.Execute(resp, finishedLookup)
		if err != nil {
			log.Println(err)
		}
	}
}
