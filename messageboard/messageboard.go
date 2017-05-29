package messageboard

import (
	"database/sql"
	"sync"
)

type MessageBoard struct {
	mutex sync.Mutex // not embedded so access to Mutex.Lock() and Mutex.Unlock() is not exported
	*sql.DB
}

func New() (mdb *MessageBoard, err error) {
	db, err := sql.Open("sqlite3", conf.MessageBoardDatabaseFilePath)
	return &MessageBoard{sync.Mutex{}, db}, err
}

func (mb *MessageBoard) lock() {
	mb.mutex.Lock()
}

func (db *MessageBoard) unlock() {
	db.mutex.Unlock()
}
