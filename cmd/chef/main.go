package main

import (
	"log"
	"net/http"

	"github.com/sauerbraten/chef/internal/collector"
	"github.com/sauerbraten/chef/internal/web"
)

func main() {
	// start collector
	coll := collector.New(conf.db, conf.ms, conf.scanInterval, conf.extraServers, conf.verbose)
	go coll.Run()

	// start web interface
	s := web.New(conf.db, conf.kidban)
	log.Println("server listening on", conf.webInterfaceAddress)
	err := http.ListenAndServe(conf.webInterfaceAddress, s)
	if err != nil {
		log.Println(err)
	}
}
