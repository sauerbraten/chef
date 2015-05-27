package main

import (
	"os"

	"github.com/sauerbraten/jsonfile"
)

type config struct {
	ServerAddress       string   `json:"irc_server_address"`
	Nick                string   `json:"irc_nick"`
	AccountName         string   `json:"irc_account_name"`
	AccountPassword     string   `json:"irc_account_password"`
	Channels            []string `json:"irc_channels"`
	TrustedUsers        []User   `json:"irc_trusted_users"`
	hostnameByUsername  map[string]string
	usernameByHostname  map[string]string
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

	initializeUsers()
}
