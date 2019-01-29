package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/sauerbraten/chef/internal/db"
	"github.com/sauerbraten/chef/pkg/ips"
	"github.com/sauerbraten/chef/pkg/kidban"
)

type WebInterface struct {
	db *db.Database

	chi.Router
	kidban *kidban.Checker
}

func NewWebInterface(db *db.Database, kidban *kidban.Checker) *WebInterface {
	r := chi.NewRouter()
	r.Use(
		middleware.RedirectSlashes,
		requestLogging,
	)

	w := &WebInterface{
		db: db,

		Router: r,
		kidban: kidban,
	}

	r.HandleFunc("/", w.frontPage())
	r.HandleFunc("/info", w.infoPage())
	r.HandleFunc("/status", w.statusPage())
	r.HandleFunc("/lookup", w.lookup())
	r.Handle("/{:[a-z]+\\.css}", http.FileServer(http.Dir("css")))

	return w

}

func requestLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		remoteAddr := req.Header.Get("X-Real-IP")
		if remoteAddr == "" {
			remoteAddr = req.RemoteAddr
		}
		log.Println(strings.Split(remoteAddr, ":")[0], "requested", req.URL.String())

		h.ServeHTTP(resp, req)
	})
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
		}{
			Status:                 w.db.Status(),
			TimeOfLastKidbanUpdate: w.kidban.TimeOfLastUpdate().UTC().Format("2006-01-02 15:04:05 MST"),
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
		nameOrIP := req.FormValue("q")

		sorting := db.ByNameFrequency
		if req.FormValue("sorting") == db.ByLastSeen.Identifier {
			sorting = db.ByLastSeen
		}

		last6MonthsOnly := !(req.FormValue("search_old") == "true")

		directLookupForced := req.FormValue("direct") == "true"

		// (permanently) redirect partial IP queries
		if ips.IsPartialOrFullCIDR(nameOrIP) {
			var subnet *net.IPNet
			subnet = ips.GetSubnet(nameOrIP)

			if nameOrIP != subnet.String() {
				u, _ := url.ParseRequestURI(req.RequestURI) // it's safe to assume this will not fail
				params := u.Query()
				params.Set("q", subnet.String())
				u.RawQuery = params.Encode()
				http.Redirect(resp, req, u.String(), http.StatusPermanentRedirect)
				return
			}
		}

		finishedLookup := w.db.Lookup(nameOrIP, sorting, last6MonthsOnly, directLookupForced)

		if req.FormValue("format") == "json" {
			err := json.NewEncoder(resp).Encode(finishedLookup)
			if err != nil {
				log.Println(err)
			}
		} else {
			err := tmpl.Execute(resp, finishedLookup)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
