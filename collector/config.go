package main

import (
	"os"
	"time"

	"github.com/sauerbraten/jsonconf"
)

type config struct {
	MasterServerAddress string        `json:"master_server_address"`
	MasterServerPort    string        `json:"master_server_port"`
	HiddenServers       []string      `json:"hidden_servers"`
	BlacklistedServers  []string      `json:"blacklisted_servers"`
	ScanIntervalSeconds time.Duration `json:"scan_interval_seconds"`
}

var conf config

func init() {
	configFilePath := "~/.chef/config.json"
	if len(os.Args) >= 2 {
		configFilePath = os.Args[1]
	}

	conf = config{}

	err := jsonconf.ParseFile(filePath, &conf)
	if err != nil {
		panic(err)
	}
}
