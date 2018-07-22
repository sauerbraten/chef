package db

import (
	"log"

	"github.com/sauerbraten/chef/ips"
)

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
func (db *Database) Lookup(nameOrIP string, sorting Sorting, directLookupForced bool) FinishedLookup {
	if ips.IsPartialOrFullCIDR(nameOrIP) {
		lowest, highest := ips.GetIpRange(ips.GetSubnet(nameOrIP))
		return FinishedLookup{
			Query:                 nameOrIP,
			InterpretedAsName:     false,
			PerformedDirectLookup: true,
			SortedByNameFrequency: sorting == ByNameFrequency,
			Results:               db.lookupIpRange(lowest, highest, sorting),
		}
	}

	return FinishedLookup{
		Query:                 nameOrIP,
		InterpretedAsName:     true,
		PerformedDirectLookup: directLookupForced,
		SortedByNameFrequency: sorting == ByNameFrequency,
		Results:               db.lookupName(nameOrIP, sorting, directLookupForced),
	}
}

func (db *Database) lookupIpRange(lowestIpInRange, highestIpInRange int64, sorting Sorting) []Sighting {
	condition := "`sightings`.`ip` in (select `rowid` from `ips` where `ip` >= ? and `ip` <= ?)"
	return db.lookup(condition, sorting, lowestIpInRange, highestIpInRange)
}

func (db *Database) lookupName(name string, sorting Sorting, directLookupForced bool) []Sighting {
	condition := "`sightings`.`name` in (select `rowid` from `names` where `name` like ?)"
	if !directLookupForced {
		condition = "`sightings`.`ip` in (select `ip` from `sightings` where " + condition + " and `ip` != (select `rowid` from `ips` where `ip` = 0))"
	}
	return db.lookup(condition, sorting, "%"+name+"%")
}

func (db *Database) lookup(condition string, sorting Sorting, args ...interface{}) []Sighting {
	const (
		columns      = "`names`.`name`, `ips`.`ip`, max(`timestamp`), `sightings`.`server`, `servers`.`ip`, `servers`.`port`, `servers`.`description`"
		joinedTables = "`sightings`, `ips` on `sightings`.`ip` = `ips`.`rowid`, `names` on `sightings`.`name` = `names`.`rowid`, `servers` on `sightings`.`server` = `servers`.`rowid`"
		grouping     = "`names`.`name`, `ips`.`ip`"
	)

	query := "select " + columns + " from " + joinedTables + " where " + condition + " group by " + grouping + " order by " + string(sorting) + " desc limit 1000"

	db.mutex.Lock()
	defer db.mutex.Unlock()

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Fatalln("error performing look up:", err)
	}
	defer rows.Close()

	return rowsToSightings(rows)
}
