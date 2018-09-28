package db

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	mutex sync.Mutex // not embedded so that access to Mutex.Lock() and Mutex.Unlock() is not exported
	*sql.DB
}

func New() (*Database, error) {
	db, err := sql.Open("sqlite3", conf.DatabaseFilePath)
	return &Database{
		mutex: sync.Mutex{},
		DB:    db,
	}, err
}
