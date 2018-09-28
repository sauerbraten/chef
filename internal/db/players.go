package db

import (
	"database/sql"
	"log"
)

// Returns the SQLite rowid of the name specified. In case no such entry exists, it is inserted and the SQLite rowid of the new entry is returned.
func (db *Database) getPlayerNameId(name string) (nameId int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	err := db.QueryRow("select `rowid` from `names` where `name` like ?", name).Scan(&nameId)
	if err == sql.ErrNoRows {
		res, err := db.Exec("insert into `names` values (?)", name)
		if err != nil {
			log.Fatal("error inserting new name:", err)
		}

		nameId, err = res.LastInsertId()
		if err != nil {
			log.Fatal("error getting ID of newly inserted name:", err)
		}
	} else if err != nil {
		log.Fatal("error getting ID of name:", err)
	}

	return
}

// Returns the SQLite rowid of the IP specified. In case no such entry exists, it is inserted and the SQLite rowid of the new entry is returned.
func (db *Database) getPlayerIpId(ip int64) (ipId int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	err := db.QueryRow("select `rowid` from `ips` where `ip` = ?", ip).Scan(&ipId)
	if err == sql.ErrNoRows {
		res, err := db.Exec("insert into `ips` values (?)", ip)
		if err != nil {
			log.Fatal("error inserting new IP:", err)
		}

		ipId, err = res.LastInsertId()
		if err != nil {
			log.Fatal("error getting ID of newly inserted IP:", err)
		}
	} else if err != nil {
		log.Fatal("error getting ID of IP:", err)
	}

	return
}
