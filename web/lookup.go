package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/sauerbraten/chef/db"
	"github.com/sauerbraten/chef/kidban"
)

func TimestampToString(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05 MST")
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

	// redirect partial IP queries
	if ips.IsPartialOrFullRange(nameOrIP) {
		var subnet *net.IPNet
		subnet, err = ips.GetSubnet(nameOrIP)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		if nameOrIP != subnet.String() {
			newURL := "/lookup?q=" + url.QueryEscape(subnet.String()) + "&sorting=" + req.FormValue("sorting")
			http.Redirect(resp, req, newURL, http.StatusTemporaryRedirect)
			return
		}
	}

	finishedLookup := storage.Lookup(nameOrIP, sorting, directLookupForced)

	if req.FormValue("plain") == "true" {
		if len(finishedLookup.Results) == 0 {
			fmt.Fprintln(resp, "nothing found!")
			return
		}

		fmt.Fprintf(resp, "%15s   %-15s   %-23s   %15s   %5s   %s\n\n", "PLAYER IP", "PLAYER NAME", "LAST SEEN", "SERVER IP", "PORT", "SERVER DESCRIPTION")

		for _, sighting := range finishedLookup.Results {
			fmt.Fprintf(resp, "%15s   %-15s   %-23s   %15s   %5d   %s\n", sighting.IP, sighting.Name, time.Unix(sighting.Timestamp, 0).UTC().Format("2006-01-02 15:04:05 MST"), sighting.ServerIP, sighting.ServerPort, sighting.ServerDescription)
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
