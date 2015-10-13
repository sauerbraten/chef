package main

import (
	"strings"
	"time"
)

func leaveMessage(senderHostname, recipientUsername, contents string) (reply string) {
	recipientHostname, ok := getHostnameByAlias(strings.ToLower(recipientUsername))
	if !ok {
		reply = "couldn't find the user '" + recipientUsername + "'!"
		return
	}

	messageBoard.LeaveMessage(senderHostname, recipientHostname, contents)

	username, _ := getUsernameByHostname(recipientHostname)
	return "your message for " + username + " was successfully stored!"
}

func getMessagesLeftForUser(recipientHostname string) (reply []string) {
	messages := messageBoard.GetMessagesLeftForUser(getMaskFromHostname(recipientHostname))
	messageBoard.DeleteMessagesLeftForUser(getMaskFromHostname(recipientHostname))

	for _, message := range messages {
		username, _ := getUsernameByHostname(message.SenderHostname)
		reply = append(reply, "("+time.Unix(message.Timestamp, 0).UTC().Format("2006-01-02 15:04:05 MST")+") "+username+": "+message.Contents)
	}

	return
}
