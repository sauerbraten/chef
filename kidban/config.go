package kidban

import (
	"os"

	"github.com/sauerbraten/jsonconf"
)

type config struct {
	KidbanRangesURL string `json:"kidban_ranges_url"`
}

var conf config

func init() {
	configFilePath := "config.json"
	if len(os.Args) >= 2 {
		configFilePath = os.Args[1]
	}

	conf = config{}

	err := jsonconf.ParseFile(configFilePath, &conf)
	if err != nil {
		panic(err)
	}

	go PeriodicallyUpdateKidbanRanges(conf.KidbanRangesURL)
}
