package db

import (
	"database/sql"
	"errors"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	mutex sync.Mutex // not embedded so that access to Mutex.Lock() and Mutex.Unlock() is not exported
	*sql.DB
}

func New(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, errors.New("db: could not open database: " + err.Error())
	}

	_, err = db.Exec("pragma foreign_keys = on")
	if err != nil {
		return nil, errors.New("db: could not enable foreign keys: " + err.Error())
	}

	return &Database{
		mutex: sync.Mutex{},
		DB:    db,
	}, nil
}
