package db

import (
	"database/sql"
	"errors"
	"strings"
)

// AddOrUpdateServer does an upsert operation to ensure a record for the server with the given information exists in the database,
// then returns that record's ID. The server's description and mod field will be updated to the given values.
// When this function returns, the server's last_seen field will be set to the current time.
func (db *Database) AddOrUpdateServer(ip string, port int, description, mod string) (serverID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	var oldDescription, oldMod string
	err := db.QueryRow("select `id`, `description`, `mod` from `servers` where `ip` = ? and `port` = ?", ip, port).Scan(&serverID, &oldDescription, &oldMod)

	switch {
	case err == nil: // server already exists
		_, err = db.Exec("update `servers` set `description` = ?, `mod` = ?, `last_seen` = strftime('%s', 'now') where `id` = ?", description, mod, serverID)
		if err != nil {
			db.fatal("update server info", err, "ip", ip, "port", port)
		}

	case errors.Is(err, sql.ErrNoRows): // new server found
		res, err := db.Exec("insert into `servers` (`ip`, `port`, `description`, `mod`) values (?, ?, ?, ?)", ip, port, description, mod)
		if err != nil {
			db.fatal("insert new server into database", err, "ip", ip, "port", port)
		}

		serverID, err = res.LastInsertId()
		if err != nil {
			db.fatal("get ID of newly inserted server", err, "ip", ip, "port", port)
		}

		// last_seen defaults to 'now' in the DB schema, so we don't need to set it

	default:
		db.fatal("get ID of server", err, "ip", ip, "port", port)
	}

	return
}

type Server struct {
	ID          int64  `json:"id"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Description string `json:"description"`
	Mod         string `json:"mod,omitempty"`
	LastSeen    int64  `json:"last_seen,omitempty"`
}

func (db *Database) FindServerByDescription(q string) []Server {
	words := strings.Fields(q)   // remove all whitespace by splitting on it
	q = strings.Join(words, "%") // place '%' SQLite wildcard between words
	q = "%" + q + "%"            // wrap in '%' wildcards

	rows, err := db.Query("select id, ip, port, description, mod, last_seen from servers where description like ? order by last_seen desc", q)
	if err != nil {
		db.fatal("find server by description", err, "server_desc_query", q)
	}
	defer rows.Close()

	results := []Server{}

	for rows.Next() {
		s := Server{}
		rows.Scan(&s.ID, &s.IP, &s.Port, &s.Description, &s.Mod, &s.LastSeen)
		results = append(results, s)
	}
	if rows.Err() != nil {
		db.fatal("iterate results of server lookup", err, "server_desc_query", q)
	}

	return results
}
