package kidban

import (
	"os"
	"time"

	"github.com/sauerbraten/jsonfile"
)

type config struct {
	KidbanRangesURL string        `json:"kidban_ranges_url"`
	UpdateInterval  time.Duration `json:"kidban_update_interval"`
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
