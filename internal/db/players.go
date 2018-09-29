package db

import (
	"database/sql"
	"log"
)

// Returns the SQLite rowid of the name specified. In case no such entry exists, it is inserted and the SQLite rowid of the new entry is returned.
func (db *Database) getPlayerNameID(name string) (nameID int64) {
	err := db.QueryRow("select `rowid` from `names` where `name` like ?", name).Scan(&nameID)
	if err == sql.ErrNoRows {
		res, err := db.Exec("insert into `names` (`name`) values (?)", name)
		if err != nil {
			log.Fatalln("error inserting new name:", err)
		}

		nameID, err = res.LastInsertId()
		if err != nil {
			log.Fatalln("error getting ID of newly inserted name:", err)
		}
	} else if err != nil {
		log.Fatalln("error getting ID of name:", err)
	}

	return
}

// Returns the SQLite rowid of the IP specified. In case no such entry exists, it is inserted and the SQLite rowid of the new entry is returned.
func (db *Database) getPlayerIpID(ip int64) (ipID int64) {
	err := db.QueryRow("select `rowid` from `ips` where `ip` = ?", ip).Scan(&ipID)
	if err == sql.ErrNoRows {
		res, err := db.Exec("insert into `ips` values (?)", ip)
		if err != nil {
			log.Fatalln("error inserting new IP:", err)
		}

		ipID, err = res.LastInsertId()
		if err != nil {
			log.Fatalln("error getting ID of newly inserted IP:", err)
		}
	} else if err != nil {
		log.Fatalln("error getting ID of IP:", err)
	}

	return
}
