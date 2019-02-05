package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sauerbraten/chef/internal/db"
	"github.com/sauerbraten/chef/pkg/kidban"
)

type WebInterface struct {
	Frontend
}

func NewWebInterface(db *db.Database, kidban *kidban.Checker) *WebInterface {
	w := &WebInterface{
		Frontend: NewFrontend(db, kidban),
	}

	w.HandleFunc("/", w.frontPage())
	w.HandleFunc("/info", w.infoPage())
	w.HandleFunc("/status", w.statusPage())
	w.HandleFunc("/lookup", w.lookup())
	w.Handle("/{:[a-z]+\\.css}", http.FileServer(http.Dir("css")))

	return w
}

func staticPageFromTemplates(files ...string) http.HandlerFunc {
	buf := new(bytes.Buffer)
	err := template.
		Must(template.ParseFiles(files...)).
		Execute(buf, nil)
	if err != nil {
		log.Fatalf("failed to build static page from template files (%s): %v\n", strings.Join(files, ", "), err)
	}

	buildTime, page := time.Now(), bytes.NewReader(buf.Bytes())

	return func(resp http.ResponseWriter, req *http.Request) {
		http.ServeContent(resp, req, "", buildTime, page)
		page.Seek(0, io.SeekStart)
	}
}

func (w *WebInterface) frontPage() http.HandlerFunc {
	return staticPageFromTemplates("templates/base.tmpl", "templates/search_form.tmpl", "templates/front.tmpl")
}

func (w *WebInterface) infoPage() http.HandlerFunc {
	return staticPageFromTemplates("templates/base.tmpl", "templates/info.tmpl")
}

func (w *WebInterface) statusPage() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("templates/base.tmpl", "templates/status.tmpl"))

	return func(resp http.ResponseWriter, req *http.Request) {
		status := struct {
			db.Status
			TimeOfLastKidbanUpdate string
			Revision               string
		}{
			Status:                 w.db.Status(),
			TimeOfLastKidbanUpdate: w.kidban.TimeOfLastUpdate().UTC().Format("2006-01-02 15:04:05 MST"),
			Revision:               gitRevision,
		}

		err := tmpl.Execute(resp, status)
		if err != nil {
			log.Println(err)
		}
	}
}

func (w *WebInterface) lookup() http.HandlerFunc {
	tmpl, err := template.
		New("base.tmpl"). // must be the base template (entry point) so templates are associated correctly by ParseFiles()
		Funcs(template.FuncMap{
			"timestring": func(timestamp int64) string { return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05") },
			"kidbanned":  func(ip string) bool { return w.kidban.IsBanned(net.ParseIP(ip)) },
		}).
		Option("missingkey=error").
		ParseFiles("templates/base.tmpl", "templates/search_form.tmpl", "templates/results.tmpl")
	if err != nil {
		log.Fatalln(err)
	}

	return func(resp http.ResponseWriter, req *http.Request) {
		w.Frontend.lookup(resp, req, func(resp http.ResponseWriter, results db.FinishedLookup) {
			err := tmpl.Execute(resp, results)
			if err != nil {
				log.Println(err)
			}
		})
	}
}
