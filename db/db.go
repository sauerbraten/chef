package db // import "github.com/sauerbraten/chef/db"

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	mutex sync.Mutex // used to lock the DB for writing access
	*sql.DB
}

func New() (*DB, error) {
	db, err := sql.Open("sqlite3", conf.FilePath)

	return &DB{sync.Mutex{}, db}, err
}

func (db *DB) lock() {
	db.mutex.Lock()
}

func (db *DB) unlock() {
	db.mutex.Unlock()
}
