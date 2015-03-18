package db

import (
	"database/sql"
	"log"
	"net"

	"github.com/sauerbraten/chef/ips"
)

type Sighting struct {
	Name              string
	IP                string
	Timestamp         int64
	ServerIP          string
	ServerPort        int
	ServerDescription string
}

// Adds an entry in the sightings table or does nothing if adding fails due to DB constraints.
func (db *DB) AddOrIgnoreSighting(name string, ip net.IP, serverId int64) {
	db.lock()
	defer db.unlock()

	_, err := db.Exec("insert or ignore into `sightings` (`name`, `ip`, `server`) values (?, ?, ?)", db.getPlayerNameId(name), db.getPlayerIpId(ips.IP2Int(ip)), serverId)
	if err != nil {
		log.Fatal("error inserting new sighting:", err)
	}
}

func rowsToSightings(rows *sql.Rows) []Sighting {
	sightings := []Sighting{}

	for rows.Next() {
		name, intIP, timestamp, serverIP, serverPort, serverDescription := "", int64(0), int64(0), "", 0, ""
		rows.Scan(&name, &intIP, &timestamp, &serverIP, &serverPort, &serverDescription)

		ip := ips.Int2IP(intIP).String()
		if ip == "255.255.255.255" {
			ip = ""
		}

		sightings = append(sightings, Sighting{
			Name:              name,
			IP:                ip,
			Timestamp:         timestamp,
			ServerIP:          serverIP,
			ServerPort:        serverPort,
			ServerDescription: serverDescription,
		})
	}

	return sightings
}
