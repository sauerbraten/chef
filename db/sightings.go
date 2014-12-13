package db

import (
	"database/sql"
	"log"
	"net"
	"regexp"
)

// checks for IP by requiring one to 4 octets. matches when there is at least the first octet. octets one to three need to end with a dot. also matches CIDR notations of ranges.
// examples:
// 123.
// 109.103.
// 11.233.109.201
// 154.93.0.0/16
var ipRegex *regexp.Regexp = regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9]?[0-9])\.((25[0-5]|2[0-4][0-9]|[01]?[0-9]?[0-9])\.)?((25[0-5]|2[0-4][0-9]|[01]?[0-9]?[0-9])\.)?(25[0-5]|2[0-4][0-9]|[01]?[0-9]?[0-9])?(\/(3[0-1]|[12]?[0-9]))?$`)

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
	ByNameFrequency         = "count(`sightings`.`name`)" // put most oftenly used name first
)

// Looks up a name or an IP or IP range (IPs are assumed to be short forms of ranges).
func (db *DB) LookUp(nameOrIP string, sorting Sorting) []Sighting {
	if ipRegex.MatchString(nameOrIP) {
		lower, upper := getIPRange(getSubnet(nameOrIP))
		return db.lookUpByIpRange(lower, upper, sorting)
	} else {
		return db.lookUpByName(nameOrIP, sorting)
	}
}

func (db *DB) lookUpByIpRange(lowerIpRangeBoundary, upperIpRangeBoundary int64, sorting Sorting) []Sighting {
	rows, err := db.Query("select `names`.`name`, `ips`.`ip`, max(`timestamp`), `servers`.`ip`, `servers`.`port`, `servers`.`description` from `sightings`, `ips` on `sightings`.`ip` = `ips`.`rowid`, `names` on `sightings`.`name` = `names`.`rowid`, `servers` on `sightings`.`server` = `servers`.`rowid` where `sightings`.`ip` in (select `rowid` from `ips` where `ip` >= ? and `ip` <= ?) group by `names`.`name`, `ips`.`ip` order by "+string(sorting)+" desc limit 1000", lowerIpRangeBoundary, upperIpRangeBoundary)
	if err != nil {
		log.Fatal("error looking up sightings by IP:", err)
	}
	defer rows.Close()

	return rowsToSightings(rows)
}

func (db *DB) lookUpByName(name string, sorting Sorting) []Sighting {
	rows, err := db.Query("select `names`.`name`, `ips`.`ip`, max(`timestamp`), `servers`.`ip`, `servers`.`port`, `servers`.`description` from `sightings`, `ips` on `sightings`.`ip` = `ips`.`rowid`, `names` on `sightings`.`name` = `names`.`rowid`, `servers` on `sightings`.`server` = `servers`.`rowid` where (`sightings`.`ip` in (select `ip` from `sightings` where `name` in (select `rowid` from `names` where `name` like ?)) and `ips`.`ip` != '') group by `names`.`name`, `ips`.`ip` order by "+string(sorting)+" desc limit 1000", "%"+name+"%")
	if err != nil {
		log.Fatal("error looking up sightings by name:", err)
	}
	defer rows.Close()

	return rowsToSightings(rows)
}

func rowsToSightings(rows *sql.Rows) []Sighting {
	sightings := []Sighting{}

	for rows.Next() {
		name, intIp, timestamp, serverIP, serverPort, serverDescription := "", int64(0), int64(0), "", 0, ""
		rows.Scan(&name, &intIp, &timestamp, &serverIP, &serverPort, &serverDescription)

		ip := intToIP(intIp).String()
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

func (db *DB) GetServerId(ip string, port int, description string) (serverId int64) {
	err := db.QueryRow("select `rowid` from `servers` where `ip` = ? and `port` = ?", ip, port).Scan(&serverId)

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
	} else {
		_, err = db.Exec("update `servers` set `description` = ? where `rowid` = ?", description, serverId)
		if err != nil {
			log.Fatal("error updating server description:", err)
		}
	}

	return
}

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

func (db *DB) AddOrIgnoreSighting(name string, ip net.IP, serverId int64) {
	_, err := db.Exec("insert or ignore into `sightings` (`name`, `ip`, `server`) values (?, ?, ?)", db.getPlayerNameId(name), db.getPlayerIpId(ipToInt(ip)), serverId)
	if err != nil {
		log.Fatal("error inserting new sighting:", err)
	}
}
