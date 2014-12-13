package main

import (
	"os"

	"github.com/sauerbraten/jsonconf"
)

type config struct {
	WebInterfaceAddress string `json:"web_interface_address"`
}

var conf config

func init() {
	filePath := ""
	if len(os.Args) < 2 {
		filePath = "config.json"
	} else {
		filePath = os.Args[1]
	}

	conf = config{}

	err := jsonconf.ParseFile(filePath, &conf)
	if err != nil {
		panic(err)
	}
}
