package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	irc "github.com/fluffle/goirc/client"
	"github.com/sauerbraten/chef/db"
	"github.com/sauerbraten/chef/messageboard"
)

var (
	storage      *db.Database
	messageBoard *messageboard.MessageBoard
	conn         *irc.Conn
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

	messageBoard, err = messageboard.New()
	if err != nil {
		log.Fatal(err)
	}
	defer messageBoard.Close()

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

	// handle joins
	conn.HandleFunc(irc.JOIN, handleJoin)

	// handle messages in channels or PM
	conn.HandleFunc(irc.PRIVMSG, handlePrivMsg)

	for {
		// tell client to connect.
		if err := conn.Connect(); err != nil {
			log.Fatal(err)
		}

		<-disconnected

		// when disconnected, wait 1 minute before trying to re-connect
		time.Sleep(1 * time.Minute)
	}
}

func handleJoin(conn *irc.Conn, line *irc.Line) {
	// only act when a trusted user joins
	if _, ok := conf.usernameByHostname[getMaskFromHostname(line.Host)]; !ok {
		return
	}

	checkMessages(line, false)
}

func handlePrivMsg(conn *irc.Conn, line *irc.Line) {
	if !strings.HasPrefix(line.Text(), ".") {
		return
	}

	if !line.Public() {
		allowed := false

		for _, user := range conf.TrustedUsers {
			if getMaskFromHostname(line.Host) == user.Host {
				allowed = true
			}
		}

		if !allowed {
			conn.Privmsg(line.Nick, "you lack access to this bot. contact "+MAINTAINER+" if you think you are eligible.")
			return
		}
	}

	firstMessageToken := strings.Split(line.Text(), " ")[0]

	switch firstMessageToken {
	case ".help", ".commands", ".about", ".usage", ".h":
		reply(line, ".names <name|IP>      – returns the five most oftenly used names by IPs that used <name> / names used by <IP>")
		reply(line, ".lastseen <name|IP>   – returns date and time a player with that <name/IP> was last seen")
		reply(line, ".message <name> <msg> - stores msg so it can be retrieved by <name> later (using the .checkmessages command)")
		reply(line, ".checkmessages        - checks if there are messages left for you and if so displays them")
	case ".names", ".name", ".nicks", ".n":
		nameOrIP := line.Text()[len(firstMessageToken)+1:]
		reply(line, nameLookup(nameOrIP))
	case ".lastseen", ".seen", ".ls", ".s":
		nameOrIP := line.Text()[len(firstMessageToken)+1:]
		reply(line, lastSeenLookup(nameOrIP))
	case ".leavemessage", ".message", ".lm", ".m":
		lineParts := strings.Split(line.Text(), " ")
		reply(line, leaveMessage(line.Host, lineParts[1], strings.Join(lineParts[2:], " ")))
	case ".checkmessages", ".cm":
		checkMessages(line, true)
	}

	// log query
	log.Println(line.Src+":", line.Text())
}

func checkMessages(line *irc.Line, explicitRequest bool) {
	messages := getMessagesLeftForUser(line.Host)

	if len(messages) == 0 {
		if explicitRequest {
			replyInPM(line.Nick, "there were no messages left for you!")
		}
		return
	} else if len(messages) == 1 {
		replyInPM(line.Nick, "there was one message left for you while you were gone:")
	} else {
		replyInPM(line.Nick, "there were "+strconv.Itoa(len(messages))+" messages left for you while you were gone:")
	}

	for _, m := range messages {
		replyInPM(line.Nick, m)
	}
}

func reply(line *irc.Line, msg string) {
	if line.Public() {
		msg = line.Nick + ": " + msg
	}

	conn.Privmsg(line.Target(), msg)
}

func replyInPM(target string, msg string) {
	conn.Privmsg(target, msg)
}
