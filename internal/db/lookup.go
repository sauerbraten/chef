package db

import (
	"encoding/json"
	"log"

	"github.com/sauerbraten/chef/pkg/ips"
)

type Sorting struct {
	Identifier  string
	DisplayName string
	sql         string
}

func (s Sorting) MarshalJSON() ([]byte, error) { return json.Marshal(s.Identifier) }

var (
	ByLastSeen = Sorting{
		Identifier:  "last_seen",
		DisplayName: "last seen",
		sql:         "`timestamp` desc", // sort most recent sighting first
	}
	ByNameFrequency = Sorting{
		Identifier:  "name_frequency",
		DisplayName: "name frequency",
		sql:         "count(`combinations`.`name`) desc", // put most oftenly used name first
	}
)

type FinishedLookup struct {
	Query                 string     `json:"query"`
	InterpretedAsName     bool       `json:"interpreted_as_name"`
	PerformedDirectLookup bool       `json:"direct"`
	Last90DaysOnly        bool       `json:"last_90_days_only"`
	Sorting               Sorting    `json:"sorting"`
	Results               []Sighting `json:"results"`
}

// Looks up a name or an IP or IP range (IPs are assumed to be short forms of ranges).
func (db *Database) Lookup(nameOrIP string, sorting Sorting, last90DaysOnly bool, directLookupForced bool) FinishedLookup {
	if ips.IsPartialOrFullCIDR(nameOrIP) {
		lowest, highest := ips.GetDecimalBoundaries(ips.GetSubnet(nameOrIP))
		return FinishedLookup{
			Query:                 nameOrIP,
			InterpretedAsName:     false,
			PerformedDirectLookup: true,
			Last90DaysOnly:        last90DaysOnly,
			Sorting:               sorting,
			Results:               db.lookupIpRange(lowest, highest, sorting, last90DaysOnly),
		}
	}

	return FinishedLookup{
		Query:                 nameOrIP,
		InterpretedAsName:     true,
		PerformedDirectLookup: directLookupForced,
		Last90DaysOnly:        last90DaysOnly,
		Sorting:               sorting,
		Results:               db.lookupName(nameOrIP, sorting, last90DaysOnly, directLookupForced),
	}
}

func (db *Database) lookupIpRange(lowestIpInRange, highestIpInRange int64, sorting Sorting, last90DaysOnly bool) []Sighting {
	condition := "`combinations`.`ip` >= ? and `combinations`.`ip` <= ?"

	if last90DaysOnly {
		condition += " and `sightings`.`timestamp` > strftime('%s', 'now', '-90 days')"
	}

	return db.lookup(condition, sorting, lowestIpInRange, highestIpInRange)
}

func (db *Database) lookupName(name string, sorting Sorting, last90DaysOnly bool, directLookupForced bool) []Sighting {
	condition := "`combinations`.`name` in (select `id` from `names` where `name` like ?)"

	if !directLookupForced {
		condition = "`combinations`.`ip` in (select `ip` from `combinations` where " + condition + " and `combinations`.`ip` != 0)"
	}

	if last90DaysOnly {
		condition += " and `sightings`.`timestamp` > strftime('%s', 'now', '-90 days')"
	}

	return db.lookup(condition, sorting, "%"+name+"%")
}

func (db *Database) lookup(condition string, sorting Sorting, args ...interface{}) []Sighting {
	const (
		columns      = "`names`.`name`, `combinations`.`ip`, max(`timestamp`), `sightings`.`server`, `servers`.`ip`, `servers`.`port`, `servers`.`description`, `servers`.`mod`"
		joinedTables = "`sightings`, `combinations` on `sightings`.`combination` = `combinations`.`id`, `names` on `combinations`.`name` = `names`.`id`, `servers` on `sightings`.`server` = `servers`.`id`"
		grouping     = "`names`.`name`, `combinations`.`ip`"
	)

	query := "select " + columns + " from " + joinedTables + " where " + condition + " group by " + grouping + " order by " + sorting.sql + " limit 1000"

	db.mutex.Lock()
	defer db.mutex.Unlock()

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Fatalln("error performing look up:", err)
	}
	defer rows.Close()

	return rowsToSightings(rows)
}
