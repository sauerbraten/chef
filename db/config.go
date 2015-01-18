package db

import (
	"os"

	"github.com/sauerbraten/jsonconf"
)

type config struct {
	FilePath string `json:"db_file_path"`
}

var conf config

func init() {
	configFilePath := "~/.chef/config.json"
	if len(os.Args) >= 2 {
		configFilePath = os.Args[1]
	}

	conf = config{}

	err := jsonconf.ParseFile(configFilePath, &conf)
	if err != nil {
		panic(err)
	}
}
