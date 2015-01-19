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

type Sorting string

const (
	ByLastSeen      Sorting = "`timestamp`"               // sort most recent sighting first
	ByNameFrequency Sorting = "count(`sightings`.`name`)" // put most oftenly used name first
)

type FinishedLookup struct {
	Query                 string
	InterpretedAsName     bool
	PerformedDirectLookup bool
	SortedByNameFrequency bool
	Results               []Sighting
}

// Looks up a name or an IP or IP range (IPs are assumed to be short forms of ranges).
func (db *DB) Lookup(nameOrIP string, sorting Sorting, directLookupForced bool) FinishedLookup {
	if ips.IsIP(nameOrIP) {
		lowest, highest := ips.GetIpRange(ips.GetSubnet(nameOrIP))
		return FinishedLookup{
			Query:                 nameOrIP,
			InterpretedAsName:     false,
			PerformedDirectLookup: true,
			SortedByNameFrequency: sorting == ByNameFrequency,
			Results:               db.lookupIpRange(lowest, highest, sorting),
		}
	} else {
		return FinishedLookup{
			Query:                 nameOrIP,
			InterpretedAsName:     true,
			PerformedDirectLookup: directLookupForced,
			SortedByNameFrequency: sorting == ByNameFrequency,
			Results:               db.lookupName(nameOrIP, sorting, directLookupForced),
		}
	}
}

func (db *DB) lookupIpRange(lowestIpInRange, highestIpInRange int64, sorting Sorting) []Sighting {
	rows, err := db.Query("select `names`.`name`, `ips`.`ip`, max(`timestamp`), `servers`.`ip`, `servers`.`port`, `servers`.`description` from `sightings`, `ips` on `sightings`.`ip` = `ips`.`rowid`, `names` on `sightings`.`name` = `names`.`rowid`, `servers` on `sightings`.`server` = `servers`.`rowid` where `sightings`.`ip` in (select `rowid` from `ips` where `ip` >= ? and `ip` <= ?) group by `names`.`name`, `ips`.`ip` order by "+string(sorting)+" desc limit 1000", lowestIpInRange, highestIpInRange)
	if err != nil {
		log.Fatal("error looking up sightings by IP:", err)
	}
	defer rows.Close()

	return rowsToSightings(rows)
}

func (db *DB) lookupName(name string, sorting Sorting, directLookupForced bool) []Sighting {
	condition := "`sightings`.`ip` in (select `ip` from `sightings` where `name` in (select `rowid` from `names` where `name` like ?) and `ip` != (select `rowid` from `ips` where `ip` = 0))"

	if directLookupForced {
		condition = "`sightings`.`name` in (select `rowid` from `names` where `name` like ?)"
	}

	rows, err := db.Query("select `names`.`name`, `ips`.`ip`, max(`timestamp`), `servers`.`ip`, `servers`.`port`, `servers`.`description` from `sightings`, `ips` on `sightings`.`ip` = `ips`.`rowid`, `names` on `sightings`.`name` = `names`.`rowid`, `servers` on `sightings`.`server` = `servers`.`rowid` where ("+condition+") group by `names`.`name`, `ips`.`ip` order by "+string(sorting)+" desc limit 1000", "%"+name+"%")
	if err != nil {
		log.Fatal("error looking up sightings by name:", err)
	}
	defer rows.Close()

	return rowsToSightings(rows)
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

// Returns the SQLite rowid of the server specified by IP and port. In case no such server exists, it is inserted and the rowid of the new entry is returned.
// If a server with that IP and port already exists but the descriptions differ, it is updated in the database.
func (db *DB) GetServerId(ip string, port int, description string) (serverId int64) {
	var descriptionInDB string
	err := db.QueryRow("select `rowid`, `description` from `servers` where `ip` = ? and `port` = ?", ip, port).Scan(&serverId, &descriptionInDB)

	if err == sql.ErrNoRows {
		res, err := db.Exec("insert into `servers` (`ip`, `port`, `description`) values (?, ?, ?)", ip, port, description)
		if err != nil {
			log.Fatal("error inserting new server into DB:", err)
		}

		serverId, err = res.LastInsertId()
		if err != nil {
			log.Fatal("error getting ID of newly inserted server:", err)
		}
	} else if err != nil {
		log.Fatal("error getting ID of server:", err)
	} else if description != descriptionInDB {
		_, err = db.Exec("update `servers` set `description` = ? where `rowid` = ?", description, serverId)
		if err != nil {
			log.Fatal("error updating server description:", err)
		}
	}

	return
}

// Returns the SQLite rowid of the name specified. In case no such entry exists, it is inserted and the rowid of the new entry is returned.
func (db *DB) getPlayerNameId(name string) (nameId int64) {
	err := db.QueryRow("select `rowid` from `names` where `name` like ?", name).Scan(&nameId)

	if err == sql.ErrNoRows {
		res, err := db.Exec("insert into `names` values (?)", name)
		if err != nil {
			log.Fatal("error inserting new name:", err)
		}

		nameId, err = res.LastInsertId()
		if err != nil {
			log.Fatal("error getting ID of newly inserted name:", err)
		}
	} else if err != nil {
		log.Fatal("error getting ID of name:", err)
	}

	return
}

// Returns the SQLite rowid of the IP specified. In case no such entry exists, it is inserted and the rowid of the new entry is returned.
func (db *DB) getPlayerIpId(ip int64) (ipId int64) {
	err := db.QueryRow("select `rowid` from `ips` where `ip` = ?", ip).Scan(&ipId)

	if err == sql.ErrNoRows {
		res, err := db.Exec("insert into `ips` values (?)", ip)
		if err != nil {
			log.Fatal("error inserting new IP:", err)
		}

		ipId, err = res.LastInsertId()
		if err != nil {
			log.Fatal("error getting ID of newly inserted IP:", err)
		}
	} else if err != nil {
		log.Fatal("error getting ID of IP:", err)
	}

	return
}

// Adds an entry in the sightings table.
func (db *DB) AddOrIgnoreSighting(name string, ip net.IP, serverId int64) {
	_, err := db.Exec("insert or ignore into `sightings` (`name`, `ip`, `server`) values (?, ?, ?)", db.getPlayerNameId(name), db.getPlayerIpId(ips.IP2Int(ip)), serverId)
	if err != nil {
		log.Fatal("error inserting new sighting:", err)
	}
}
