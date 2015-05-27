package messageboard

import "log"

type Message struct {
	SenderHostname    string
	RecipientHostname string
	Contents          string
	Timestamp         int64
}

// Retrieves all messages still left to read for the user with the specified hostmask
func (mb *MessageBoard) GetMessagesLeftForUser(recipientHostname string) []Message {
	rows, err := mb.Query("select `sender_hostname`, `contents`, `timestamp` from `messages` where `recipient_hostname` = ? order by `timestamp`", recipientHostname)
	if err != nil {
		log.Fatal("error retreiving messages left for user:", err)
	}
	defer rows.Close()

	messages := []Message{}

	for rows.Next() {
		message := Message{RecipientHostname: recipientHostname}
		rows.Scan(&message.SenderHostname, &message.Contents, &message.Timestamp)
		messages = append(messages, message)
	}

	return messages
}

// Deletes all messages still left to read for the user with the specified hostmask
func (mb *MessageBoard) DeleteMessagesLeftForUser(recipientHostname string) {
	_, err := mb.Exec("delete from `messages` where `recipient_hostname` = ?", recipientHostname)
	if err != nil {
		log.Fatal("error deleting messages left for user:", err)
	}
}

// Stores a message for the recipient to read later
func (mb *MessageBoard) LeaveMessage(senderHostname, recipientHostname, contents string) {
	_, err := mb.Exec("insert or ignore into `messages` (`sender_hostname`, `recipient_hostname`, `contents`) values (?, ?, ?)", senderHostname, recipientHostname, contents)
	if err != nil {
		log.Fatal("error inserting new message into database:", err)
	}
}
