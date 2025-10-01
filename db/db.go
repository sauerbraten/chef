package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Database struct {
	reader *sql.DB
	writer *sql.DB
}

func New(path string) (*Database, error) {
	writer, err := sql.Open("sqlite3", "file:"+path+"?_txlock=immediate")
	if err != nil {
		return nil, fmt.Errorf("db: open database: %w", err)
	}
	err = setPragmas(writer)
	if err != nil {
		return nil, err
	}
	writer.SetMaxOpenConns(1)

	reader, err := sql.Open("sqlite3", "file:"+path+"?_txlock=immediate")
	if err != nil {
		return nil, fmt.Errorf("db: open database: %w", err)
	}
	err = setPragmas(reader)
	if err != nil {
		return nil, err
	}
	reader.SetMaxOpenConns(max(4, runtime.NumCPU()))

	return &Database{
		reader: reader,
		writer: writer,
	}, nil
}

func setPragmas(db *sql.DB) error {
	_, err := db.Exec("pragma foreign_keys = on")
	if err != nil {
		return fmt.Errorf("db: enable foreign keys: %w", err)
	}
	_, err = db.Exec("pragma journal_mode = wal")
	if err != nil {
		return fmt.Errorf("db: enable WAL: %w", err)
	}
	_, err = db.Exec("pragma synchronous = normal")
	if err != nil {
		return fmt.Errorf("db: set synchronous=normal: %w", err)
	}
	_, err = db.Exec("pragma busy_timeout = 5000")
	if err != nil {
		return fmt.Errorf("db: set busy_timeout=5000: %w", err)
	}
	return nil
}

func (db *Database) fatal(err error, args ...any) {
	slog.Error(err.Error(), args...)
	slog.Error("exiting after fatal DB error")
	// we deal with errors by turning off (and letting the supervisor turn us on again)
	db.reader.Close()
	db.writer.Close()
	os.Exit(1)
}

func (db *Database) inTx(fn func(tx *sql.Tx) error) (err error) {
	tx, err := db.writer.Begin()
	if err != nil {
		return fmt.Errorf("db: open tx: %w", err)
	}

	defer func() {
		if err == nil {
			return
		}
		rollbackErr := tx.Rollback() // no-op after successful commit
		if rollbackErr != nil {
			rollbackErr = fmt.Errorf("db: rollback tx: %w", rollbackErr)
			db.fatal(errors.Join(err, rollbackErr))
		}
	}()

	err = fn(tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("db: commit tx: %w", err)
	}

	return nil
}

func (db *Database) Close() error {
	err := db.reader.Close()
	if err != nil {
		return fmt.Errorf("db: close connection: %w", err)
	}
	err = db.writer.Close()
	if err != nil {
		return fmt.Errorf("db: close connection: %w", err)
	}
	return nil
}
