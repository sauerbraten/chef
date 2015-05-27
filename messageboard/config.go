package messageboard

import (
	"os"

	"github.com/sauerbraten/jsonfile"
)

type config struct {
	MessageBoardDatabaseFilePath string `json:"mb_db_file_path"`
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
