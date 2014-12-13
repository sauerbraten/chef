package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func New() (*DB, error) {
	db, err := sql.Open("sqlite3", conf.FilePath)

	return &DB{db}, err
}

func (db *DB) Close() {
	db.DB.Close()
}
