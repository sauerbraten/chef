package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	s, err := NewServer()
	if err != nil {
		log.Println(err)
	}

	r := chi.NewRouter()
	r.Use(
		middleware.RedirectSlashes,
		requestLogging,
	)

	r.HandleFunc("/", s.frontPage())
	r.HandleFunc("/info", s.infoPage())
	r.HandleFunc("/status", s.statusPage())
	r.HandleFunc("/lookup", s.lookup())
	r.Handle("/{:[a-z]+\\.css}", http.FileServer(http.Dir("css")))

	// start listening
	log.Println("server listening on", conf.WebInterfaceAddress)
	err = http.ListenAndServe(conf.WebInterfaceAddress, r)
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
