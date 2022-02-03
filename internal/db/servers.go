package db

import (
	"database/sql"
	"log"
	"strings"
)

type Server struct {
	ID          int64  `json:"id"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Description string `json:"description"`
	Mod         string `json:"mod,omitempty"`
	LastSeen    int64  `json:"last_seen,omitempty"`
}

// Returns the ID of the server specified by IP and port. In case no such server exists, it is inserted and the rowid of the new entry is returned.
// If a server with that IP and port already exists but the description or mod changed, the entry is updated in the database.
func (db *Database) GetServerID(ip string, port int, description string, mod int8, protocol int) (serverID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	var (
		oldDescription string
		oldMod         int8
	)
	err := db.QueryRow("select `id`, `description`, `mod` from `servers` where `ip` = ? and `port` = ? and `protocol` = ?", ip, port, protocol).Scan(&serverID, &oldDescription, &oldMod)

	if err == sql.ErrNoRows {
		res, err := db.Exec("insert into `servers` (`ip`, `port`, `description`, `mod`, `protocol`) values (?, ?, ?, ?, ?)", ip, port, description, mod, protocol)
		if err != nil {
			log.Fatalln("error inserting new server into database:", err)
		}

		serverID, err = res.LastInsertId()
		if err != nil {
			log.Fatalln("error getting ID of newly inserted server:", err)
		}
	} else if err != nil {
		log.Fatalln("error getting ID of server:", err)
	} else if description != oldDescription || mod != oldMod {
		_, err = db.Exec("update `servers` set `description` = ?, `mod` = ? where `id` = ?", description, mod, serverID)
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

func (db *Database) FindServerByDescription(desc string) []Server {
	desc = strings.Replace(desc, " ", "", -1)         // remove spaces
	desc = strings.Join(strings.Split(desc, ""), "%") // place '%' SQLite wildcard between characters
	desc = "%" + desc + "%"                           // wrap in '%' wildcards

	db.mutex.Lock()
	defer db.mutex.Unlock()

	rows, err := db.Query("select `id`, `ip`, `port`, `description`, `mod`, `last_seen` from `servers` where `description` like ?", desc)
	if err != nil {
		log.Fatalln("error finding server by description:", err)
	}
	defer rows.Close()

	results := []Server{}

	for rows.Next() {
		s := Server{}
		rows.Scan(&s.ID, &s.IP, &s.Port, &s.Description, &s.Mod, &s.LastSeen)
		results = append(results, s)
	}

	return results
}
