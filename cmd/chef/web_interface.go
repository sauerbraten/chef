package main

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sauerbraten/chef/internal/db"
)

var (
	//go:embed templates css
	embedded  embed.FS
	templates = func() fs.FS { f, _ := fs.Sub(embedded, "templates"); return f }()
	css       = func() fs.FS { f, _ := fs.Sub(embedded, "css"); return f }()
)

type WebInterface struct {
	Frontend
}

func NewWebInterface(db *db.Database) *WebInterface {
	w := &WebInterface{
		Frontend: NewFrontend(db),
	}

	w.HandleFunc("/", w.frontPage())
	w.HandleFunc("/info", w.infoPage())
	w.HandleFunc("/status", w.statusPage())
	w.HandleFunc("/sightings", w.sightings())
	w.HandleFunc("/games", w.games())
	w.HandleFunc("/games/{id}", w.game())
	w.Handle("/{:[a-z]+\\.css}", http.FileServer(http.FS(css)))

	return w
}

func (w *WebInterface) staticPageFromTemplates(files ...string) http.HandlerFunc {
	buf := new(bytes.Buffer)
	err := template.
		Must(template.ParseFS(templates, files...)).
		Execute(buf, nil)
	if err != nil {
		log.Fatalf("failed to build static page from template files (%s): %v\n", strings.Join(files, ", "), err)
	}

	buildTime, page := time.Now(), bytes.NewReader(buf.Bytes())

	return func(resp http.ResponseWriter, req *http.Request) {
		http.ServeContent(resp, req, "dummy.html", buildTime, page)
		page.Seek(0, io.SeekStart)
	}
}

func (w *WebInterface) frontPage() http.HandlerFunc {
	return w.staticPageFromTemplates("base.tmpl", "sightings_search_form.tmpl", "games_search_form.tmpl", "front.tmpl")
}

func (w *WebInterface) infoPage() http.HandlerFunc {
	return w.staticPageFromTemplates("base.tmpl", "info.tmpl")
}

func (w *WebInterface) statusPage() http.HandlerFunc {
	tmpl := template.
		New("base.tmpl"). // must be the base template (entry point) so templates are associated correctly by ParseFiles()
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"formatInt": func(i int) string {
				s := strconv.Itoa(i)
				if i < 1000 {
					return s
				}
				f, s := s[len(s)-3:], s[:len(s)-3]
				for len(s) > 3 {
					f = s[len(s)-3:] + "," + f
					s = s[:len(s)-3]
				}
				return s + "," + f
			},
		})
	tmpl, err := tmpl.ParseFS(templates, "base.tmpl", "status.tmpl")
	if err != nil {
		log.Fatalln(err)
	}

	return func(resp http.ResponseWriter, req *http.Request) {
		status := struct {
			db.Status
			Revision string
		}{
			Status:   w.db.Status(),
			Revision: gitRevision,
		}

		err := tmpl.Execute(resp, status)
		if err != nil {
			log.Println(err)
		}
	}
}

func (w *WebInterface) sightings() http.HandlerFunc {
	tmpl := template.
		New("base.tmpl"). // must be the base template (entry point) so templates are associated correctly by ParseFS()
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"timestring": func(timestamp int64) string { return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05") },
		})
	tmpl, err := tmpl.ParseFS(templates, "base.tmpl", "sightings_search_form.tmpl", "sightings_results.tmpl")
	if err != nil {
		log.Fatalln(err)
	}

	return func(resp http.ResponseWriter, req *http.Request) {
		w.Frontend.lookupSightings(resp, req, func(resp http.ResponseWriter, results db.FinishedLookup) {
			err := tmpl.Execute(resp, results)
			if err != nil {
				log.Println(err)
			}
		})
	}
}

func (w *WebInterface) games() http.HandlerFunc {
	tmpl := template.
		New("base.tmpl"). // must be the base template (entry point) so templates are associated correctly by ParseFS()
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"timestring": func(timestamp int64) string { return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05") },
		})
	tmpl, err := tmpl.ParseFS(templates, "base.tmpl", "games_search_form.tmpl", "games_results.tmpl")
	if err != nil {
		log.Fatalln(err)
	}

	return func(resp http.ResponseWriter, req *http.Request) {
		w.Frontend.lookupGames(resp, req, func(resp http.ResponseWriter, result db.FinishedGamesLookup) {
			err := tmpl.Execute(resp, result)
			if err != nil {
				log.Println(err)
			}
		})
	}
}

func (w *WebInterface) game() http.HandlerFunc {
	tmpl := template.
		New("base.tmpl"). // must be the base template (entry point) so templates are associated correctly by ParseFS()
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"timestring": func(timestamp int64) string { return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05") },
		})
	tmpl, err := tmpl.ParseFS(templates, "base.tmpl", "games_search_form.tmpl", "game.tmpl")
	if err != nil {
		log.Fatalln(err)
	}

	return func(resp http.ResponseWriter, req *http.Request) {
		w.Frontend.fetchGame(resp, req, func(resp http.ResponseWriter, result GameWithStats) {
			err := tmpl.Execute(resp, result)
			if err != nil {
				log.Println(err)
			}
		})
	}
}
