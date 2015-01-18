package main

import (
	"log"
	"net/http"
	"strings"
	"text/template"
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

	r.HandleFunc("/", front)
	r.HandleFunc("/lookup", lookupSightings)
	r.HandleFunc("/status", status)
	r.Handle("/{fn:[a-z]+\\.css}", http.FileServer(http.Dir("css")))

	// start listening
	log.Println("server listening on", conf.WebInterfaceAddress)
	err = http.ListenAndServe(conf.WebInterfaceAddress, r)
	if err != nil {
		log.Println(err)
	}
}

func front(resp http.ResponseWriter, req *http.Request) {
	http.ServeFile(resp, req, "html/front.html")
}

type Results struct {
	Query     string
	Sightings []db.Sighting
}

func TimestampToString(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05 MST")
}

func lookupSightings(resp http.ResponseWriter, req *http.Request) {
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

	results := Results{
		Query:     nameOrIP,
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

func status(resp http.ResponseWriter, req *http.Request) {
	logRequest(req)

	status := storage.Status()

	statusTempl := template.Must(template.ParseFiles("html/status.html"))
	statusTempl.Execute(resp, status)
}

func logRequest(req *http.Request) {
	log.Println(strings.Split(req.RemoteAddr, ":")[0], "requested", req.URL.String())
}
