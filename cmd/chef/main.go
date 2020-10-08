package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const gitRevision = "<filled in by CI service>"

func main() {
	// start collector
	coll := NewCollector(conf.db, conf.ms, conf.scanInterval, conf.extraServers, conf.verbose)
	go coll.Run()

	// start web frontends
	r := chi.NewRouter()
	r.Use(
		middleware.RedirectSlashes,
		requestLogging,
	)
	r.Mount("/api", NewAPI(conf.db))
	r.Mount("/", NewWebInterface(conf.db, conf.kidban))
	log.Println("server listening on", conf.webInterfaceAddress)
	err := http.ListenAndServe(conf.webInterfaceAddress, r)
	if err != nil {
		log.Println(err)
	}
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
