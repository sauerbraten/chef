package main

import (
	"errors"
	"os"
	"time"

	"github.com/sauerbraten/jsonfile"

	"github.com/sauerbraten/chef/internal/db"
	"github.com/sauerbraten/chef/pkg/kidban"
	"github.com/sauerbraten/chef/pkg/master"
)

type config struct {
	db *db.Database

	ms           *master.Server
	extraServers []string
	scanInterval time.Duration
	verbose      bool

	webInterfaceAddress string
	kidban              *kidban.Checker
}

var conf config

func init() {
	configFilePath := "config.json"
	if len(os.Args) >= 2 {
		configFilePath = os.Args[1]
	}

	_conf := struct {
		DatabaseFilePath string `json:"db_file_path"`

		MasterServerAddress string   `json:"master_server_address"`
		ExtraServers        []string `json:"extra_servers"`
		ScanInterval        string   `json:"scan_interval"`
		Verbose             bool     `json:"verbose"`

		WebInterfaceAddress  string `json:"web_interface_address"`
		KidbanRangesURL      string `json:"kidban_ranges_url"`
		KidbanUpdateInterval string `json:"kidban_update_interval"`
	}{}

	err := jsonfile.ParseFile(configFilePath, &_conf)
	if err != nil {
		panic(err)
	}

	_db, err := db.New(_conf.DatabaseFilePath)
	if err != nil {
		panic(errors.New("database initialization failed: " + err.Error()))
	}

	ms := master.New(_conf.MasterServerAddress, 15*time.Second)

	scanInterval, err := time.ParseDuration(_conf.ScanInterval)
	if err != nil {
		panic(errors.New("parsing scan interval failed: " + err.Error()))
	}

	kidbanUpdateInterval, err := time.ParseDuration(_conf.KidbanUpdateInterval)
	if err != nil {
		panic(errors.New("parsing kidban refresh interval failed: " + err.Error()))
	}

	kidban, err := kidban.NewChecker(_conf.KidbanRangesURL, kidbanUpdateInterval)
	if err != nil {
		panic(errors.New("kidban checker initialization failed: " + err.Error()))
	}

	conf = config{
		db: _db,

		ms:           ms,
		extraServers: _conf.ExtraServers,
		scanInterval: scanInterval,
		verbose:      _conf.Verbose,

		webInterfaceAddress: _conf.WebInterfaceAddress,
		kidban:              kidban,
	}
}
