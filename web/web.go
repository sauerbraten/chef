package main

import (
	"log"
	"net/http"
	"strings"
	"text/template"

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
	r.HandleFunc("/lookup", lookup)
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
	logRequest(req)
	http.ServeFile(resp, req, "html/front.html")
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
