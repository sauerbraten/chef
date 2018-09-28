package db

import (
	"os"

	"github.com/sauerbraten/jsonfile"
)

type config struct {
	DatabaseFilePath string `json:"db_file_path"`
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
