package db

import (
	"database/sql"
	"log"
	"net"

	"github.com/sauerbraten/chef/pkg/ips"
)

type Combination struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

// Returns the ID of the name-IP combination specified. In case no such entry exists, it is inserted and the ID of the new entry is returned.
func (db *Database) GetCombinationID(name string, ip net.IP) (id int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	nameID, ipID := db.getNameID(name), ips.IP2Int(ip)
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
	err := db.QueryRow("select `rowid` from `names` where `name` = ?", name).Scan(&nameID)
	if err == sql.ErrNoRows {
		res, err := db.Exec("insert into `names` (`name`) values (?)", name)
		if err != nil {
			log.Fatalln("error inserting new name:", err, "name:", name)
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
