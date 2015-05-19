package db // import "github.com/sauerbraten/chef/db"

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	mutex sync.Mutex // used to lock the database for writing access
	*sql.DB
}

func New() (*Database, error) {
	db, err := sql.Open("sqlite3", conf.FilePath)

	return &Database{sync.Mutex{}, db}, err
}

func (db *Database) lock() {
	db.mutex.Lock()
}

func (db *Database) unlock() {
	db.mutex.Unlock()
}
