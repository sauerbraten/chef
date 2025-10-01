package db

import (
	"database/sql"
	"fmt"
)

type Status struct {
	NamesCount        int
	IPsCount          int
	CombinationsCount int
	SightingsCount    int
	ServersCount      int
}

func (db *Database) Status() (status Status) {
	err := db.inTx(func(tx *sql.Tx) error {
		err := tx.QueryRow("select count(*) from names").Scan(&status.NamesCount)
		if err != nil {
			return fmt.Errorf("get names count for status: %w", err)
		}

		err = tx.QueryRow("select count(distinct ip) from combinations where ip != 0").Scan(&status.IPsCount)
		if err != nil {
			return fmt.Errorf("get IPs count for status: %w", err)
		}

		err = tx.QueryRow("select count(*) from combinations where ip != 0").Scan(&status.CombinationsCount)
		if err != nil {
			return fmt.Errorf("get combinations count for status: %w", err)
		}

		err = tx.QueryRow("select count(*) from sightings").Scan(&status.SightingsCount)
		if err != nil {
			return fmt.Errorf("get sightings count for status: %w", err)
		}

		err = tx.QueryRow("select count(*) from servers").Scan(&status.ServersCount)
		if err != nil {
			return fmt.Errorf("get servers count for status: %w", err)
		}

		return nil
	})
	if err != nil {
		db.fatal(err)
	}

	return
}
