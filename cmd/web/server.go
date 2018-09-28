package main

import (
	"bytes"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/sauerbraten/chef/internal/db"
	"github.com/sauerbraten/chef/internal/kidban"
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

func staticPageFromTemplates(files ...string) http.HandlerFunc {
	buf := new(bytes.Buffer)
	err := template.
		Must(template.ParseFiles(files...)).
		Option("missingkey=error").
		Execute(buf, nil)
	if err != nil {
		log.Fatalf("failed to build static page from template files (%s): %v\n", strings.Join(files, ", "), err)
	}

	buildTime, page := time.Now(), bytes.NewReader(buf.Bytes())

	return func(resp http.ResponseWriter, req *http.Request) {
		http.ServeContent(resp, req, "", buildTime, page)
	}
}

func (s *server) frontPage() http.HandlerFunc {
	return staticPageFromTemplates("templates/base.tmpl", "templates/front.tmpl")
}

func (s *server) infoPage() http.HandlerFunc {
	return staticPageFromTemplates("templates/base.tmpl", "templates/info.tmpl")
}

func (s *server) statusPage() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("templates/base.tmpl", "templates/status.tmpl"))

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
