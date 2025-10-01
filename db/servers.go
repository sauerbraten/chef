package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// AddOrUpdateServer does an upsert operation to ensure a record for the server with the given information exists in the database,
// then returns that record's ID. The server's description and mod field will be updated to the given values.
// When this function returns, the server's last_seen field will be set to the current time.
func (db *Database) AddOrUpdateServer(ip string, port int, description, mod string) (serverID int64) {
	err := db.inTx(func(tx *sql.Tx) (err error) {
		serverID, err = db.addOrUpdateServer(tx, ip, port, description, mod)
		return
	})
	if err != nil {
		db.fatal(err, "ip", ip, "port", port)
	}
	return
}

func (db *Database) addOrUpdateServer(tx *sql.Tx, ip string, port int, description, mod string) (int64, error) {
	var (
		id             int64
		oldDescription string
		oldMod         string
	)
	err := tx.QueryRow("select `id`, `description`, `mod` from `servers` where `ip` = ? and `port` = ?", ip, port).Scan(&id, &oldDescription, &oldMod)

	switch {
	case err == nil: // server already exists
		_, err = tx.Exec("update `servers` set `description` = ?, `mod` = ?, `last_seen` = strftime('%s', 'now') where `id` = ?", description, mod, id)
		if err != nil {
			return -1, fmt.Errorf("update server info: %w", err)
		}
		return id, nil

	case errors.Is(err, sql.ErrNoRows): // new server found
		res, err := tx.Exec("insert into `servers` (`ip`, `port`, `description`, `mod`) values (?, ?, ?, ?)", ip, port, description, mod)
		if err != nil {
			return -1, fmt.Errorf("insert new server into database: %w", err)
		}
		id, err = res.LastInsertId()
		if err != nil {
			return -1, fmt.Errorf("get ID of newly inserted server: %w", err)
		}
		// last_seen defaults to 'now' in the DB schema, so we don't need to set it
		return id, nil

	default:
		return -1, fmt.Errorf("get ID of server: %w", err)
	}
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
	q = strings.Join(words, "%") // place '%' (SQL wildcard character) between words
	q = "%" + q + "%"            // wrap in '%' wildcards

	rows, err := db.reader.Query("select id, ip, port, description, mod, last_seen from servers where description like ? order by last_seen desc", q)
	if err != nil {
		db.fatal(fmt.Errorf("find server by description: %w", err), "server_desc_query", q)
	}
	defer rows.Close()

	results := []Server{}
	for rows.Next() {
		s := Server{}
		rows.Scan(&s.ID, &s.IP, &s.Port, &s.Description, &s.Mod, &s.LastSeen)
		results = append(results, s)
	}
	if rows.Err() != nil {
		db.fatal(fmt.Errorf("iterate results of server lookup: %w", err), "server_desc_query", q)
	}

	return results
}
