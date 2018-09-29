package main

import (
	"net"
	"os"

	"github.com/sauerbraten/jsonfile"
)

type config struct {
	DatabaseFilePath    string         `json:"db_file_path"`
	MasterServerAddress string         `json:"master_server_address"`
	ExtraServers        []string       `json:"extra_servers"`
	extraServers        []*net.UDPAddr // to hold addresses after resolving them at initialization
	ScanInterval        string         `json:"scan_interval"`
	Verbose             bool           `json:"verbose"`
}

var conf config

func init() {
	configFilePath := "config.json"
	if len(os.Args) >= 2 {
		configFilePath = os.Args[1]
	}

	conf = config{}

	err := jsonfile.ParseFile(configFilePath, &conf)
	if err != nil {
		panic(err)
	}
}
