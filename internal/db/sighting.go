package db

import (
	"database/sql"
	"log"
	"net"

	"github.com/sauerbraten/chef/pkg/ips"
)

type Server struct {
	ID          int64  `json"id"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Description string `json:"description"`
}

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

	_, err := db.Exec("insert or ignore into `sightings` (`name`, `ip`, `server`) values (?, ?, ?)", db.getPlayerNameID(name), db.getPlayerIpID(ips.IP2Int(ip)), serverID)
	if err != nil {
		log.Fatalln("error inserting new sighting:", err)
	}
}

func rowsToSightings(rows *sql.Rows) []Sighting {
	sightings := []Sighting{}

	for rows.Next() {
		name, intIP, timestamp, serverID, serverIP, serverPort, serverDescription := "", int64(0), int64(0), int64(0), "", 0, ""
		rows.Scan(&name, &intIP, &timestamp, &serverID, &serverIP, &serverPort, &serverDescription)

		ip := ips.Int2IP(intIP).String()
		if ip == "255.255.255.255" {
			ip = ""
		}

		sightings = append(sightings, Sighting{
			Name:      name,
			IP:        ip,
			Timestamp: timestamp,
			Server: Server{
				ID:          serverID,
				IP:          serverIP,
				Port:        serverPort,
				Description: serverDescription,
			},
		})
	}

	return sightings
}