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

func (db *Database) lookupIpRange(lowestIpInRange, highestIpInRange int64, sorting Sorting) []Sighting {
	rows, err := db.Query("select `names`.`name`, `ips`.`ip`, max(`timestamp`), `servers`.`ip`, `servers`.`port`, `servers`.`description` from `sightings`, `ips` on `sightings`.`ip` = `ips`.`rowid`, `names` on `sightings`.`name` = `names`.`rowid`, `servers` on `sightings`.`server` = `servers`.`rowid` where `sightings`.`ip` in (select `rowid` from `ips` where `ip` >= ? and `ip` <= ?) group by `names`.`name`, `ips`.`ip` order by "+string(sorting)+" desc limit 1000", lowestIpInRange, highestIpInRange)
	if err != nil {
		log.Fatal("error looking up sightings by IP:", err)
	}
	defer rows.Close()

	return rowsToSightings(rows)
}

func (db *Database) lookupName(name string, sorting Sorting, directLookupForced bool) []Sighting {
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
