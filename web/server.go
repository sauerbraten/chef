package main

import (
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/sauerbraten/chef/db"
	"github.com/sauerbraten/chef/kidban"
)

type server struct {
	db *db.Database
}

func NewServer() (*server, error) {
	storage, err := db.New()
	if err != nil {
		return nil, errors.New("could not create server: " + err.Error())
	}

	return &server{
		db: storage,
	}, nil
}

func (s *server) frontPage(resp http.ResponseWriter, req *http.Request) {
	http.ServeFile(resp, req, "html/front.html")
}

func (s *server) infoPage(resp http.ResponseWriter, req *http.Request) {
	http.ServeFile(resp, req, "html/info.html")
}

func (s *server) statusPage() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("html/status.html"))

	return func(resp http.ResponseWriter, req *http.Request) {
		status := struct {
			db.Status
			TimeOfLastKidbanUpdate string
		}{
			Status:                 s.db.Status(),
			TimeOfLastKidbanUpdate: kidban.GetTimeOfLastUpdate().UTC().Format("2006-01-02 15:04:05 MST"),
		}

		err := tmpl.Execute(resp, status)
		if err != nil {
			log.Println(err)
		}
	}
}
