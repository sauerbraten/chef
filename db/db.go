package db

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	mutex sync.Mutex // not embedded sp access to Mutex.Lock() and Mutex.Unlock() is not exported
	*sql.DB
}

func New() (*Database, error) {
	db, err := sql.Open("sqlite3", conf.DatabaseFilePath)
	return &Database{sync.Mutex{}, db}, err
}

func (db *Database) lock() {
	db.mutex.Lock()
}

func (db *Database) unlock() {
	db.mutex.Unlock()
}
