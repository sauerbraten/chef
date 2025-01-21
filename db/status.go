package db

type Status struct {
	NamesCount        int
	IPsCount          int
	CombinationsCount int
	SightingsCount    int
	ServersCount      int
}

func (db *Database) Status() (status Status) {
	err := db.QueryRow("select count(*) from names").Scan(&status.NamesCount)
	if err != nil {
		db.fatal("get names count for status", err)
	}

	err = db.QueryRow("select count(distinct ip) from combinations where ip != 0").Scan(&status.IPsCount)
	if err != nil {
		db.fatal("get IPs count for status", err)
	}

	err = db.QueryRow("select count(*) from combinations where ip != 0").Scan(&status.CombinationsCount)
	if err != nil {
		db.fatal("get combinations count for status", err)
	}

	err = db.QueryRow("select count(*) from sightings").Scan(&status.SightingsCount)
	if err != nil {
		db.fatal("get sightings count for status", err)
	}

	err = db.QueryRow("select count(*) from servers").Scan(&status.ServersCount)
	if err != nil {
		db.fatal("get servers count for status", err)
	}

	return
}
