package db

import (
	"database/sql"
	"log"
	"strings"
)

// Returns the ID of the server specified by IP and port. In case no such server exists, it is inserted and the rowid of the new entry is returned.
// If a server with that IP and port already exists but the descriptions differ, the description is updated in the database.
func (db *Database) GetServerID(ip string, port int, description string) (serverID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	var descriptionInDatabase string
	err := db.QueryRow("select `id`, `description` from `servers` where `ip` = ? and `port` = ?", ip, port).Scan(&serverID, &descriptionInDatabase)

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
		_, err = db.Exec("update `servers` set `description` = ? where `id` = ?", description, serverID)
		if err != nil {
			log.Fatalln("error updating server description:", err)
		}
	}

	return
}

// Updates the 'last seen' timestamp of a server.
func (db *Database) UpdateServerLastSeen(serverID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, err := db.Exec("update `servers` set `last_seen` = strftime('%s', 'now') where `id` = ?", serverID)
	if err != nil {
		log.Fatalln("error updating server's 'last seen' timestamp:", err)
	}
}

type Server struct {
	ID          int64  `json:"id"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Description string `json:"description"`
	LastSeen    int64  `json:"last_seen,omitempty"`
}

func (db *Database) FindServerByDescription(desc string) (results []Server) {
	desc = strings.Replace(desc, " ", "", -1)         // remove spaces
	desc = strings.Join(strings.Split(desc, ""), "%") // place '%' SQLite wildcard between characters
	desc = "%" + desc + "%"                           // wrap in '%' wildcards

	db.mutex.Lock()
	defer db.mutex.Unlock()

	rows, err := db.Query("select `id`, `ip`, `port`, `description`, `last_seen` from `servers` where `description` like ?", desc)
	if err != nil {
		log.Fatalln("error finding server by description:", err)
	}
	defer rows.Close()

	for rows.Next() {
		s := Server{}
		rows.Scan(&s.ID, &s.IP, &s.Port, &s.Description, &s.LastSeen)
		results = append(results, s)
	}

	return
}
