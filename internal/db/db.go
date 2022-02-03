package db

import (
	"database/sql"
	"errors"
	"strings"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	_ "modernc.org/sqlite"
)

type Database struct {
	mutex sync.Mutex // not embedded so that access to Mutex.Lock() and Mutex.Unlock() is not exported
	*sqlx.DB
}

func New(filepath string) (*Database, error) {
	db, err := sqlx.Open("sqlite", filepath)
	if err != nil {
		return nil, errors.New("db: opening database: " + err.Error())
	}

	db.Mapper = reflectx.NewMapperFunc("json", strings.ToLower) // use json struct tags

	_, err = db.Exec("pragma foreign_keys = on")
	if err != nil {
		db.Close()
		return nil, errors.New("db: enabling foreign keys: " + err.Error())
	}

	err = migrateUp(db.DB, filepath)
	if err != nil {
		db.Close()
		return nil, err
	}

	_, err = db.Exec("pragma journal_mode = wal")
	if err != nil {
		db.Close()
		return nil, errors.New("db: enabling WAL mode: " + err.Error())
	}

	return &Database{
		mutex: sync.Mutex{},
		DB:    db,
	}, nil
}

func migrateUp(db *sql.DB, filepath string) error {
	srcFiles := &file.File{}
	src, err := srcFiles.Open("file://migrations")
	if err != nil {
		return errors.New("db: opening migration source files: " + err.Error())
	}
	defer src.Close()

	mDB, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return errors.New("db: using database for migrations: " + err.Error())
	}

	m, err := migrate.NewWithInstance("./migrations/", src, filepath, mDB)
	if err != nil {
		return errors.New("db: setting up migrations: " + err.Error())
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return errors.New("db: migrating database: " + err.Error())
	}

	return nil
}
