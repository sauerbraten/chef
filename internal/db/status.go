package db

import "log"

type Status struct {
	NamesCount        int
	IPsCount          int
	CombinationsCount int
	SightingsCount    int
	GamesCount        int
	ServersCount      int
}

func (db *Database) Status() (status Status) {
	err := db.QueryRow("select count(*) from names").Scan(&status.NamesCount)
	if err != nil {
		log.Fatalln("getting names count:", err)
	}

	err = db.QueryRow("select count(distinct ip) from combinations where ip != 0").Scan(&status.IPsCount)
	if err != nil {
		log.Fatalln("getting IPs count:", err)
	}

	err = db.QueryRow("select count(*) from combinations where ip != 0").Scan(&status.CombinationsCount)
	if err != nil {
		log.Fatalln("getting combinations count:", err)
	}

	err = db.QueryRow("select count(*) from sightings").Scan(&status.SightingsCount)
	if err != nil {
		log.Fatalln("getting sightings count:", err)
	}

	err = db.QueryRow("select count(*) from games").Scan(&status.GamesCount)
	if err != nil {
		log.Fatalln("getting games count:", err)
	}

	err = db.QueryRow("select count(*) from servers").Scan(&status.ServersCount)
	if err != nil {
		log.Fatalln("getting servers count:", err)
	}

	return
}
