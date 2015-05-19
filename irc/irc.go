package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	irc "github.com/fluffle/goirc/client"
	"github.com/sauerbraten/chef/db"
)

var (
	storage *db.Database
	conn    *irc.Conn
)

const (
	MAINTAINER string = "pix"
)

func main() {
	var err error
	storage, err = db.New()
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	ircConfig := irc.NewConfig(conf.Nick)
	ircConfig.Me.Ident = "chef"
	ircConfig.Me.Name = "pix' spy bot"
	ircConfig.Server = conf.ServerAddress
	ircConfig.NewNick = func(n string) string { return n + "_" }

	disconnected := make(chan bool)

	conn = irc.Client(ircConfig)

	conn.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		if conf.AccountName != "" && conf.AccountPassword != "" {
			conn.Privmsg("AuthServ@Services.GameSurge.net", fmt.Sprintf("auth %s %s", conf.AccountName, conf.AccountPassword))
			time.Sleep(500 * time.Millisecond)
		}

		for _, channel := range conf.Channels {
			conn.Join(channel)
		}
	})

	conn.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		disconnected <- true
	})

	conn.HandleFunc(irc.PRIVMSG, handlePrivMsg)

	for {
		// Tell client to connect.
		if err := conn.Connect(); err != nil {
			log.Fatal(err)
		}

		<-disconnected

		// when disconnected, wait 1 minute before trying to re-connect
		time.Sleep(1 * time.Minute)
	}
}

func handlePrivMsg(conn *irc.Conn, line *irc.Line) {
	if !strings.HasPrefix(line.Text(), ".") {
		return
	}

	if !line.Public() {
		allowed := false

		for _, user := range conf.TrustedUsers {
			if line.Host == user {
				allowed = true
			}
		}

		if !allowed {
			conn.Privmsg(line.Nick, "you lack access to this bot. contact "+MAINTAINER+" if you think you are eligible.")
			return
		}
	}

	firstMessageToken := strings.Split(line.Text(), " ")[0]

	if isHelpCommandAlias(firstMessageToken) {
		reply(line, ".names <name|IP>    – returns the five most oftenly used names by IPs that used <name> / names used by <IP>")
		reply(line, ".lastseen <name|IP> – returns date and time a player with that <name/IP> was last seen")
	} else if isNameLookupCommandAlias(firstMessageToken) {
		nameOrIP := line.Text()[len(firstMessageToken)+1:]
		reply(line, nameLookup(nameOrIP))
	} else if isLastSeenLookupCommandAlias(firstMessageToken) {
		nameOrIP := line.Text()[len(firstMessageToken)+1:]
		reply(line, lastSeenLookup(nameOrIP))
	}

	// log query
	log.Println(line.Src+":", line.Text())
}

func isIncluded(s string, slice []string) bool {
	for _, element := range slice {
		if element == s {
			return true
		}
	}

	return false
}

func isHelpCommandAlias(s string) bool {
	helpAliases := []string{".help", ".commands", ".about", ".usage", ".h"}

	return isIncluded(s, helpAliases)
}

func isNameLookupCommandAlias(s string) bool {
	nameLookupAliases := []string{".names", ".name", ".nicks", ".n"}

	return isIncluded(s, nameLookupAliases)
}

func isLastSeenLookupCommandAlias(s string) bool {
	lastSeenLookupAliases := []string{".lastseen", ".seen", ".ls", ".s"}

	return isIncluded(s, lastSeenLookupAliases)
}

func reply(line *irc.Line, msg string) {
	if line.Public() {
		msg = line.Nick + ": " + msg
	}

	conn.Privmsg(line.Target(), msg)
}
