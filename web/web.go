package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sauerbraten/chef/db"
	"github.com/sauerbraten/chef/kidban"
)

var storage *db.Database

func main() {
	var err error
	storage, err = db.New()
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	r := mux.NewRouter()
	r.StrictSlash(true)

	r.HandleFunc("/", frontPage)
	r.HandleFunc("/lookup", lookup)
	r.HandleFunc("/status", statusPage)
	r.HandleFunc("/info", infoPage)
	r.Handle("/{fn:[a-z]+\\.css}", http.FileServer(http.Dir("css")))

	// start listening
	log.Println("server listening on", conf.WebInterfaceAddress)
	err = http.ListenAndServe(conf.WebInterfaceAddress, r)
	if err != nil {
		log.Println(err)
	}
}

func frontPage(resp http.ResponseWriter, req *http.Request) {
	logRequest(req)
	http.ServeFile(resp, req, "html/front.html")
}

type status struct {
	db.Status
	TimeOfLastKidbanUpdate string
}

func statusPage(resp http.ResponseWriter, req *http.Request) {
	logRequest(req)

	status := &status{
		Status:                 storage.Status(),
		TimeOfLastKidbanUpdate: kidban.GetTimeOfLastUpdate().UTC().Format("2006-01-02 15:04:05 MST"),
	}

	statusTempl := template.Must(template.ParseFiles("html/status.html"))
	err := statusTempl.Execute(resp, status)

	if err != nil {
		log.Println(err)
	}
}

func infoPage(resp http.ResponseWriter, req *http.Request) {
	logRequest(req)
	http.ServeFile(resp, req, "html/info.html")
}

func logRequest(req *http.Request) {
	log.Println(strings.Split(req.RemoteAddr, ":")[0], "requested", req.URL.String())
}
