package main

import (
	"log"
	"net/http"
)

const gitRevision = "<filled in by CI service>"

func main() {
	// start collector
	coll := NewCollector(conf.db, conf.ms, conf.scanInterval, conf.extraServers, conf.verbose)
	go coll.Run()

	// start web interface
	w := NewWebInterface(conf.db, conf.kidban)
	log.Println("server listening on", conf.webInterfaceAddress)
	err := http.ListenAndServe(conf.webInterfaceAddress, w)
	if err != nil {
		log.Println(err)
	}
}
