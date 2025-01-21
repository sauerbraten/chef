package chef

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/sauerbraten/chef/db"
)

var (
	//go:embed templates css
	embedded embed.FS

	templates fs.FS = func() fs.FS { f, _ := fs.Sub(embedded, "templates"); return f }()
	css       fs.FS = func() fs.FS { f, _ := fs.Sub(embedded, "css"); return f }()

	// for status page
	gitRevision = func() string {
		info, ok := debug.ReadBuildInfo()
		if ok {
			for _, bs := range info.Settings {
				if bs.Key == "vcs.revision" {
					return bs.Value[:min(len(bs.Value), 8)]
				}
			}
		}
		return "unknown"
	}()
)

type WebUI struct {
	http.Handler
	db *db.Database
}

func NewWebUI(db *db.Database) *WebUI {
	wui := &WebUI{
		db: db,
	}

	r := http.NewServeMux()
	r.HandleFunc("GET /", wui.frontPage())
	r.HandleFunc("GET /info", wui.infoPage())
	r.HandleFunc("GET /status", wui.statusPage())
	r.HandleFunc("GET /lookup", wui.lookup())
	r.Handle("GET /css/", http.StripPrefix("/css/", http.FileServer(http.FS(css))))

	wui.Handler = r

	return wui
}

func (wui *WebUI) staticPageFromTemplates(files ...string) http.HandlerFunc {
	buf := new(bytes.Buffer)
	err := template.
		Must(template.ParseFS(templates, files...)).
		Execute(buf, nil)
	if err != nil {
		slog.Error("build static page from template files", "files", files, "erorr", err)
		os.Exit(1)
	}

	buildTime, page := time.Now(), bytes.NewReader(buf.Bytes())

	return func(resp http.ResponseWriter, req *http.Request) {
		http.ServeContent(resp, req, "dummy.html", buildTime, page)
		page.Seek(0, io.SeekStart)
	}
}

func (wui *WebUI) frontPage() http.HandlerFunc {
	return wui.staticPageFromTemplates("base.tmpl", "search_form.tmpl", "front.tmpl")
}

func (wui *WebUI) infoPage() http.HandlerFunc {
	return wui.staticPageFromTemplates("base.tmpl", "info.tmpl")
}

func (wui *WebUI) statusPage() http.HandlerFunc {
	tmpl := template.
		New("base.tmpl"). // must be the base template (entry point) so templates are associated correctly by ParseFS()
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
			Status:   wui.db.Status(),
			Revision: gitRevision,
		}

		err := tmpl.Execute(resp, status)
		if err != nil {
			log.Println(err)
		}
	}
}

func (wui *WebUI) lookup() http.HandlerFunc {
	tmpl := template.
		New("base.tmpl"). // must be the base template (entry point) so templates are associated correctly by ParseFS()
		Option("missingkey=error").
		Funcs(template.FuncMap{
			"timestring": func(timestamp int64) string { return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05") },
		})
	tmpl, err := tmpl.ParseFS(templates, "base.tmpl", "search_form.tmpl", "results.tmpl")
	if err != nil {
		log.Fatalln(err)
	}

	return func(resp http.ResponseWriter, req *http.Request) {
		nameOrIP, sorting, last90DaysOnly, directLookupForced, redirected := parseLookupRequest(resp, req)
		if redirected {
			return
		}

		results := wui.db.Lookup(nameOrIP, sorting, last90DaysOnly, directLookupForced)

		err := tmpl.Execute(resp, results)
		if err != nil {
			log.Println(err)
		}
	}
}
