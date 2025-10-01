package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net"

	"github.com/sauerbraten/chef/ips"
)

// Returns the ID of the name-IP combination specified. In case no such entry exists, it is inserted and the ID of the new entry is returned.
func (db *Database) getCombinationID(tx *sql.Tx, name string, ip net.IP) (id int64, err error) {
	nameID, err := db.getNameID(tx, name)
	if err != nil {
		return -1, err
	}

	err = tx.QueryRow("select `id` from `combinations` where `name` = ? and `ip` = ?", nameID, ips.IP2Int(ip)).Scan(&id)

	switch {
	case err == nil: // combination already exists
		return

	case errors.Is(err, sql.ErrNoRows): // new combination found
		res, err := tx.Exec("insert or ignore into `combinations` (`name`, `ip`) values (?, ?)", nameID, ips.IP2Int(ip))
		if err != nil {
			return -1, fmt.Errorf("insert new name-IP combination: %w", err)
		}
		id, err = res.LastInsertId()
		if err != nil {
			return -1, fmt.Errorf("get ID of newly inserted name-IP combination: %w", err)
		}
		return id, nil

	default:
		return -1, fmt.Errorf("get ID of name-IP combination: %w", err)
	}
}

// Returns the ID of the name specified. In case no such entry exists, it is inserted and the ID of the new entry is returned.
func (db *Database) getNameID(tx *sql.Tx, name string) (int64, error) {
	var id int64
	err := tx.QueryRow("select `id` from `names` where `name` like ?", name).Scan(&id)

	switch {
	case err == nil: // name already exists
		return id, nil

	case errors.Is(err, sql.ErrNoRows): // new name found
		res, err := tx.Exec("insert into `names` (`name`) values (?)", name)
		if err != nil {
			return -1, fmt.Errorf("insert new name: %w", err)
		}
		id, err = res.LastInsertId()
		if err != nil {
			return -1, fmt.Errorf("get ID of newly inserted name: %w", err)
		}
		return id, nil

	default:
		return -1, fmt.Errorf("get ID of name: %w", err)
	}
}
