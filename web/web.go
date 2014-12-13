package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sauerbraten/chef/db"
)

var storage *db.DB

func main() {
	var err error
	storage, err = db.New()
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/{sorting:names|lastseen}/{nameOrIp:.*}", lookUp)
	r.HandleFunc("/status", status)
	r.HandleFunc("/{*}", usage)
	r.HandleFunc("/", usage)

	// start listening
	log.Println("server listening on", conf.WebInterfaceAddress)
	err = http.ListenAndServe(conf.WebInterfaceAddress, r)
	if err != nil {
		log.Println(err)
	}
}

func usage(resp http.ResponseWriter, req *http.Request) {
	http.ServeFile(resp, req, "./web_usage.txt")
	return
}

func lookUp(resp http.ResponseWriter, req *http.Request) {
	logRequest(req)

	vars := mux.Vars(req)
	nameOrIp := vars["nameOrIp"]

	var sorting db.Sorting
	switch vars["sorting"] {
	case "names":
		sorting = db.ByNameFrequency
	case "lastseen":
		sorting = db.ByLastSeen
	}

	sightings := storage.LookUp(nameOrIp, sorting)

	if len(sightings) == 0 {
		fmt.Fprintln(resp, "nothing found!")
		return
	}

	fmt.Fprintf(resp, "%15s   %-15s   %-23s   %15s   %5s   %s\n\n", "PLAYER IP", "PLAYER NAME", "LAST SEEN", "SERVER IP", "PORT", "SERVER DESCRIPTION")

	for _, sighting := range sightings {
		fmt.Fprintf(resp, "%15s   %-15s   %-23s   %15s   %5d   %s\n", sighting.IP, sighting.Name, time.Unix(sighting.Timestamp, 0).UTC().Format("2006-01-02 15:04:05 MST"), sighting.ServerIP, sighting.ServerPort, sighting.ServerDescription)
	}
}

func status(resp http.ResponseWriter, req *http.Request) {
	logRequest(req)

	status := storage.Status()

	fmt.Fprintln(resp, "DATABASE STATUS")
	fmt.Fprintln(resp)
	fmt.Fprintf(resp, "%7d names\n", status.NamesCount)
	fmt.Fprintf(resp, "%7d IPs\n", status.IPsCount)
	fmt.Fprintf(resp, "%7d combinations of name and IP\n", status.CombinationsCount)
	fmt.Fprintf(resp, "%7d sightings\n", status.SightingsCount)
	fmt.Fprintf(resp, "%7d servers\n", status.ServersCount)
	fmt.Fprintln(resp)
	fmt.Fprintln(resp, "all numbers are distinct (unique) counts")
}

func logRequest(req *http.Request) {
	log.Println(strings.Split(req.RemoteAddr, ":")[0], "requested", req.URL.String())
}
