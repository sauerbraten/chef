package main

import (
	"log"
	"strings"
	"time"

	"github.com/sauerbraten/chef/db"
	"github.com/sauerbraten/chef/ips"
	irc "github.com/thoj/go-ircevent"
)

var (
	storage *db.DB
	conn    *irc.Connection
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

	conn = irc.IRC(conf.Nick, conf.Nick)

	err = conn.Connect("irc.gamesurge.net:6667")
	if err != nil {
		log.Fatal(err)
	}

	conn.AddCallback("001", func(e *irc.Event) {
		if conf.AccountName != "" && conf.AccountPassword != "" {
			conn.Privmsgf("AuthServ@Services.GameSurge.net", "auth %s %s", conf.AccountName, conf.AccountPassword)
			time.Sleep(500 * time.Millisecond)
		}

		for _, channel := range conf.Channels {
			conn.Join(channel)
		}
	})

	/* verbose output for debugging
	conn.AddCallback("NOTICE", func(e *irc.Event) {
		log.Println(e.Message())
	})

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		log.Println(e.Message())
	})
	*/

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		if !strings.HasPrefix(e.Message(), ".") {
			return
		}

		if isPM(e) {
			allowed := false

			for _, user := range conf.TrustedUsers {
				if e.Host == user {
					allowed = true
				}
			}

			if !allowed {
				conn.Privmsg(e.Nick, "you lack access to this bot. contact "+MAINTAINER+" if you think you are eligible.")
				return
			}
		}

		firstMessageToken := strings.Split(e.Message(), " ")[0]

		if isHelpCommandAlias(firstMessageToken) {
			reply(e, ".names <name|IP>    – returns the five most oftenly used names by IPs that used <name> / names used by <IP>")
			reply(e, ".lastseen <name|IP> – returns date and time a player with that <name/IP> was last seen")
		} else if isNameLookUpCommandAlias(firstMessageToken) {
			nameOrIP := e.Message()[len(firstMessageToken)+1:]
			reply(e, nameLookUp(nameOrIP))
		} else if isLastSeenLookUpCommandAlias(firstMessageToken) {
			nameOrIP := e.Message()[len(firstMessageToken)+1:]
			reply(e, lastSeenLookUp(nameOrIP))
		}

		// log query
		log.Println(e.Source+":", e.Message())
	})

	conn.Loop()
}

func isPM(e *irc.Event) bool {
	return e.Arguments[0] == conf.Nick
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

func isNameLookUpCommandAlias(s string) bool {
	nameLookupAliases := []string{".names", ".name", ".nicks", ".n"}

	return isIncluded(s, nameLookupAliases)
}

func isLastSeenLookUpCommandAlias(s string) bool {
	lastSeenLookupAliases := []string{".lastseen", ".seen", ".ls", ".s"}

	return isIncluded(s, lastSeenLookupAliases)
}

func reply(e *irc.Event, msg string) {
	target := e.Arguments[0]

	if isPM(e) {
		target = e.Nick
	} else {
		msg = e.Nick + ": " + msg
	}

	conn.Privmsg(target, msg)
}

func sanitize(s string) string {
	// don't replace '/' in IP ranges
	if !ips.IsIP(s) {
		return strings.NewReplacer("/", "_", "?", "%3F").Replace(s)
	} else {
		return s
	}
}
