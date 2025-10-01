package db

import (
	"database/sql"
	"fmt"
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
	err := db.inTx(func(tx *sql.Tx) (err error) {
		return db.addOrIgnoreSighting(tx, name, ip, serverID)
	})
	if err != nil {
		db.fatal(err, "name", name, "ip", ip, "server_id", serverID)
	}
}

func (db *Database) addOrIgnoreSighting(tx *sql.Tx, name string, ip net.IP, serverID int64) error {
	combinationID, err := db.getCombinationID(tx, name, ip)
	if err != nil {
		return fmt.Errorf("insert new sighting: %w", err)
	}
	_, err = tx.Exec("insert or ignore into `sightings` (`combination`, `server`) values (?, ?)", combinationID, serverID)
	if err != nil {
		return fmt.Errorf("insert new sighting: %w", err)
	}
	return nil
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
