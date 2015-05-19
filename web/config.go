package main

import (
	"os"

	"github.com/sauerbraten/jsonfile"
)

type config struct {
	WebInterfaceHostname           string `json:"web_interface_hostname"`
	WebInterfaceInternalListenport string `json:"web_interface_internal_listenport"`
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
