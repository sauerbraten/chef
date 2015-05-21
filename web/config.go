package main

import (
	"os"

	"github.com/sauerbraten/jsonfile"
)

type config struct {
	WebInterfaceAddress string `json:"web_interface_address"`
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
