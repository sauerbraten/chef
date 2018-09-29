package db

import "log"

type Status struct {
	NamesCount        int
	IPsCount          int
	CombinationsCount int
	SightingsCount    int
	ServersCount      int
}

func (db *Database) Status() (status Status) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	err := db.QueryRow("select count(*) from `names`").Scan(&status.NamesCount)
	if err != nil {
		log.Fatalln("error getting names count:", err)
	}

	err = db.QueryRow("select count(*) from `ips`").Scan(&status.IPsCount)
	if err != nil {
		log.Fatalln("error getting IPs count:", err)
	}

	err = db.QueryRow("select count(*) from (select distinct `name`, `ip` from `sightings`)").Scan(&status.CombinationsCount)
	if err != nil {
		log.Fatalln("error getting combinations count:", err)
	}

	err = db.QueryRow("select count(*) from `sightings`").Scan(&status.SightingsCount)
	if err != nil {
		log.Fatalln("error getting sightings count:", err)
	}

	err = db.QueryRow("select count(*) from `servers`").Scan(&status.ServersCount)
	if err != nil {
		log.Fatalln("error getting servers count:", err)
	}

	return
}
