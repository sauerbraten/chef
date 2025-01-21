package db

import (
	"database/sql"
	"errors"
	"net"

	"github.com/sauerbraten/chef/ips"
)

// Returns the ID of the name-IP combination specified. In case no such entry exists, it is inserted and the ID of the new entry is returned.
func (db *Database) getCombinationID(name string, ip net.IP) (id int64) {
	nameID, ipID := db.getNameID(name), ips.IP2Int(ip)
	err := db.QueryRow("select `id` from `combinations` where `name` = ? and `ip` = ?", nameID, ipID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		res, err := db.Exec("insert or ignore into `combinations` (`name`, `ip`) values (?, ?)", nameID, ipID)
		if err != nil {
			db.fatal("insert new name-IP combination", err, "name", name, "ip", ip)
		}

		id, err = res.LastInsertId()
		if err != nil {
			db.fatal("get ID of newly inserted name-IP combination", err, "name", name, "ip", ip)
		}
	} else if err != nil {
		db.fatal("get ID of name-IP combination", err, "name", name, "ip", ip)
	}

	return
}

// Returns the ID of the name specified. In case no such entry exists, it is inserted and the ID of the new entry is returned.
func (db *Database) getNameID(name string) (nameID int64) {
	err := db.QueryRow("select `id` from `names` where `name` like ?", name).Scan(&nameID)
	if errors.Is(err, sql.ErrNoRows) {
		res, err := db.Exec("insert into `names` (`name`) values (?)", name)
		if err != nil {
			db.fatal("insert new name", err, "name", name)
		}

		nameID, err = res.LastInsertId()
		if err != nil {
			db.fatal("get ID of newly inserted name", err, "name", name)
		}
	} else if err != nil {
		db.fatal("get ID of name", err, "name", name)
	}

	return
}
