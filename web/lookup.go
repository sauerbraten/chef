package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/sauerbraten/chef/db"
)

type Lookup struct {
	Query                 string
	SortedByNameFrequency bool
	DirectLookup          bool
}

type Results struct {
	Lookup    Lookup
	Sightings []db.Sighting
}

func TimestampToString(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05 MST")
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

	sightings := storage.Lookup(nameOrIP, sorting, directLookupForced)

	if req.FormValue("plain") == "true" {
		if len(sightings) == 0 {
			fmt.Fprintln(resp, "nothing found!")
			return
		}

		fmt.Fprintf(resp, "%15s   %-15s   %-23s   %15s   %5s   %s\n\n", "PLAYER IP", "PLAYER NAME", "LAST SEEN", "SERVER IP", "PORT", "SERVER DESCRIPTION")

		for _, sighting := range sightings {
			fmt.Fprintf(resp, "%15s   %-15s   %-23s   %15s   %5d   %s\n", sighting.IP, sighting.Name, time.Unix(sighting.Timestamp, 0).UTC().Format("2006-01-02 15:04:05 MST"), sighting.ServerIP, sighting.ServerPort, sighting.ServerDescription)
		}
	} else {
		results := Results{
			Lookup: Lookup{
				Query: nameOrIP,
				SortedByNameFrequency: sorting == db.ByNameFrequency,
				DirectLookup:          directLookupForced,
			},
			Sightings: sightings,
		}

		resultsTempl := template.New("results.html")
		resultsTempl = resultsTempl.Funcs(template.FuncMap{"timestring": TimestampToString})
		resultsTempl = template.Must(resultsTempl.ParseFiles("html/results.html"))
		err := resultsTempl.Execute(resp, results)
		if err != nil {
			log.Println(err)
		}
	}
}
