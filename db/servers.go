package db

import (
	"database/sql"
	"log"
)

// Returns the SQLite rowid of the server specified by IP and port. In case no such server exists, it is inserted and the rowid of the new entry is returned.
// If a server with that IP and port already exists but the descriptions differ, the description is updated in the database.
func (db *Database) GetServerId(ip string, port int, description string) (serverId int64) {
	var descriptionInDatabase string
	err := db.QueryRow("select `rowid`, `description` from `servers` where `ip` = ? and `port` = ?", ip, port).Scan(&serverId, &descriptionInDatabase)

	if err == sql.ErrNoRows {
		db.lock()
		defer db.unlock()

		res, err := db.Exec("insert into `servers` (`ip`, `port`, `description`) values (?, ?, ?)", ip, port, description)
		if err != nil {
			log.Fatal("error inserting new server into database:", err)
		}

		serverId, err = res.LastInsertId()
		if err != nil {
			log.Fatal("error getting ID of newly inserted server:", err)
		}
	} else if err != nil {
		log.Fatal("error getting ID of server:", err)
	} else if description != descriptionInDatabase {
		db.lock()
		defer db.unlock()

		_, err = db.Exec("update `servers` set `description` = ? where `rowid` = ?", description, serverId)
		if err != nil {
			log.Fatal("error updating server description:", err)
		}
	}

	return
}
