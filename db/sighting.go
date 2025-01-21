package db

import (
	"database/sql"
	"log/slog"
	"net"

	"github.com/sauerbraten/chef/ips"
)

type Sighting struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Timestamp int64  `json:"time_seen"`
	Server    Server `json:"server"`
}

// Adds an entry in the sightings table or does nothing if adding fails due to database constraints.
func (db *Database) AddOrIgnoreSighting(name string, ip net.IP, serverID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, err := db.Exec("insert or ignore into `sightings` (`combination`, `server`) values (?, ?)", db.getCombinationID(name, ip), serverID)
	if err != nil {
		db.fatal("insert new sighting", err, "name", name, "server_id", serverID)
	}
}

func rowsToSightings(rows *sql.Rows) []Sighting {
	sightings := []Sighting{}

	for rows.Next() {
		sighting, intIP := Sighting{}, int64(0)
		srv := &sighting.Server

		err := rows.Scan(&sighting.Name, &intIP, &sighting.Timestamp, &srv.ID, &srv.IP, &srv.Port, &srv.Description, &srv.Mod)
		if err != nil {
			slog.Error("db: scan DB row into sighting", "error", err)
		}

		sighting.IP = ips.Int2IP(intIP).String()
		if sighting.IP == "255.255.255.255" {
			sighting.IP = ""
		}

		sightings = append(sightings, sighting)
	}

	return sightings
}
