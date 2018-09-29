package db

import (
	"database/sql"
	"log"
)

// Returns the SQLite rowid of the server specified by IP and port. In case no such server exists, it is inserted and the rowid of the new entry is returned.
// If a server with that IP and port already exists but the descriptions differ, the description is updated in the database.
func (db *Database) GetServerID(ip string, port int, description string) (serverID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	var descriptionInDatabase string
	err := db.QueryRow("select `rowid`, `description` from `servers` where `ip` = ? and `port` = ?", ip, port).Scan(&serverID, &descriptionInDatabase)

	if err == sql.ErrNoRows {
		res, err := db.Exec("insert into `servers` (`ip`, `port`, `description`) values (?, ?, ?)", ip, port, description)
		if err != nil {
			log.Fatalln("error inserting new server into database:", err)
		}

		serverID, err = res.LastInsertId()
		if err != nil {
			log.Fatalln("error getting ID of newly inserted server:", err)
		}
	} else if err != nil {
		log.Fatalln("error getting ID of server:", err)
	} else if description != descriptionInDatabase {
		_, err = db.Exec("update `servers` set `description` = ? where `rowid` = ?", description, serverID)
		if err != nil {
			log.Fatalln("error updating server description:", err)
		}
	}

	return
}
