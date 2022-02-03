package db

import (
	"database/sql"
	"log"

	"github.com/sauerbraten/chef/pkg/ips"
)

type Sighting struct {
	Combination
	SeenAt int64  `json:"seen_at"`
	Server Server `json:"server"`
}

// Adds an entry in the sightings table or does nothing if adding fails due to database constraints.
func (db *Database) AddOrIgnoreSighting(combinationID, serverID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, err := db.Exec("insert or ignore into `sightings` (`combination`, `server`) values (?, ?)", combinationID, serverID)
	if err != nil {
		log.Fatalln("error inserting new sighting:", err)
	}
}

func rowsToSightings(rows *sql.Rows) []Sighting {
	sightings := []Sighting{}

	for rows.Next() {
		sighting, intIP := Sighting{}, int64(0)
		srv := &sighting.Server

		err := rows.Scan(&sighting.Name, &intIP, &sighting.SeenAt, &srv.ID, &srv.IP, &srv.Port, &srv.Description, &srv.Mod)
		if err != nil {
			log.Println("error scanning DB row into sighting:", err)
		}

		sighting.IP = ips.Int2IP(intIP).String()
		if sighting.IP == "255.255.255.255" {
			sighting.IP = ""
		}

		sightings = append(sightings, sighting)
	}

	return sightings
}
