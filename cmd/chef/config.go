package main

import (
	"errors"
	"os"
	"time"

	"github.com/sauerbraten/jsonfile"

	"github.com/sauerbraten/chef/internal/db"
	"github.com/sauerbraten/chef/pkg/master"
)

type config struct {
	db *db.Database

	ms              *master.Server
	refreshInterval time.Duration
	scanInterval    time.Duration
	verbose         bool

	webInterfaceAddress string
}

var conf config

func init() {
	configFilePath := "config.json"
	if len(os.Args) >= 2 {
		configFilePath = os.Args[1]
	}

	_conf := struct {
		DatabaseFilePath string `json:"db_file_path"`

		MasterServerAddress string `json:"master_server_address"`
		RefreshInterval     string `json:"refresh_interval"`
		ScanInterval        string `json:"scan_interval"`
		Verbose             bool   `json:"verbose"`

		WebInterfaceAddress string `json:"web_interface_address"`
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

	refreshInterval, err := time.ParseDuration(_conf.RefreshInterval)
	if err != nil {
		panic(errors.New("parsing refresh interval failed: " + err.Error()))
	}

	scanInterval, err := time.ParseDuration(_conf.ScanInterval)
	if err != nil {
		panic(errors.New("parsing scan interval failed: " + err.Error()))
	}

	conf = config{
		db: _db,

		ms:              ms,
		refreshInterval: refreshInterval,
		scanInterval:    scanInterval,
		verbose:         _conf.Verbose,

		webInterfaceAddress: _conf.WebInterfaceAddress,
	}
}
