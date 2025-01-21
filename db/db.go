package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"sync"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Database struct {
	mutex *sync.Mutex // not embedded so that access to Mutex.Lock() and Mutex.Unlock() is not exported
	*sql.DB
}

func New(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("db: open database: %w", err)
	}

	_, err = db.Exec("pragma foreign_keys = on")
	if err != nil {
		return nil, fmt.Errorf("db: enable foreign keys: %w", err)
	}

	_, err = db.Exec("pragma journal_mode = wal")
	if err != nil {
		return nil, fmt.Errorf("db: enable WAL: %w", err)
	}

	return &Database{
		mutex: new(sync.Mutex),
		DB:    db,
	}, nil
}

func (db *Database) fatal(msg string, err error, args ...any) {
	slog.Error("db: "+msg, append(args, "error", err)...)
	// we deal with errors by turning off (and letting the supervisor turn us on again)
	db.Close()
	os.Exit(1)
}
