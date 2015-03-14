package main

import (
	"net"
	"os"
	"time"

	"github.com/sauerbraten/jsonfile"
)

type config struct {
	MasterServerAddress string          `json:"master_server_address"`
	MasterServerPort    string          `json:"master_server_port"`
	ExtraServers        []string        `json:"extra_servers"`
	extraServers        []*net.UDPAddr  // to hold addresses after resolving them at initialization
	GreylistedServers   []string        `json:"greylisted_servers"`
	greylistedServers   map[string]bool // map (basically a set) for easier lookups
	ScanIntervalSeconds time.Duration   `json:"scan_interval_seconds"`
	Verbose             bool            `json:"verbose"`
}

var conf config

func init() {
	configFilePath := "config.json"
	if len(os.Args) >= 2 {
		configFilePath = os.Args[1]
	}

	conf = config{
		greylistedServers: map[string]bool{},
	}

	err := jsonfile.ParseFile(configFilePath, &conf)
	if err != nil {
		panic(err)
	}
}
