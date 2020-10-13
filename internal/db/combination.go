package db

import (
	"database/sql"
	"log"
	"net"

	"github.com/sauerbraten/chef/pkg/ips"
)

// Returns the ID of the name-IP combination specified. In case no such entry exists, it is inserted and the ID of the new entry is returned.
func (db *Database) getCombinationID(name string, ip net.IP) (id int64) {
	nameID, ipID := db.getNameID(name), db.getIpID(ips.IP2Int(ip))
	err := db.QueryRow("select `id` from `combinations` where `name` = ? and `ip` = ?", nameID, ipID).Scan(&id)
	if err == sql.ErrNoRows {
		res, err := db.Exec("insert or ignore into `combinations` (`name`, `ip`) values (?, ?)", nameID, ipID)
		if err != nil {
			log.Fatalln("error inserting new name-IP combination:", err)
		}

		id, err = res.LastInsertId()
		if err != nil {
			log.Fatalln("error getting ID of newly inserted name-IP combination:", err)
		}
	} else if err != nil {
		log.Fatalln("error getting ID of name-IP combination:", err)
	}

	return
}

// Returns the ID of the name specified. In case no such entry exists, it is inserted and the ID of the new entry is returned.
func (db *Database) getNameID(name string) (nameID int64) {
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
func (db *Database) getIpID(ip int64) (ipID int64) {
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
