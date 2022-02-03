package db

import (
	"log"

	"github.com/sauerbraten/chef/pkg/ips"
)

type Stats struct {
	Combination
	GameID     int64  `json:"game"`
	Team       string `json:"team"`
	Frags      int    `json:"frags"`
	Deaths     int    `json:"deaths"`
	Accuracy   int    `json:"accuracy"`
	Teamkills  int    `json:"teamkills"`
	Flags      int    `json:"flags"`
	RecordedAt int64  `json:"recorded_at"`
}

func (db *Database) AddOrUpdateStats(combinationID, gameID int64, team string, frags, deaths, accuracy, teamkills, flags int) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, err := db.Exec(
		"insert into stats (combination, game, team, frags, deaths, accuracy, teamkills, flags) values (?, ?, ?, ?, ?, ?, ?, ?)",
		combinationID, gameID, team, frags, deaths, accuracy, teamkills, flags)
	if err != nil {
		log.Fatalf("inserting stats (comb %v, game %v): %v\n", combinationID, gameID, err)
	}
}

func (db *Database) GetGameStats(id int64) []Stats {
	const (
		columns      = "names.name, combinations.ip, team, frags, deaths, accuracy, teamkills, flags, max(recorded_at)"
		joinedTables = "stats, combinations on stats.combination = combinations.id, names on combinations.name = names.id"
		condition    = "game = ?"
		grouping     = "combination"
	)

	query := "select " + columns + " from " + joinedTables + " where " + condition + "group by " + grouping

	db.mutex.Lock()
	defer db.mutex.Unlock()

	rows, err := db.Query(query, id)
	if err != nil {
		log.Fatalf("fetching stats from DB: %s: %v\n", query, err)
	}

	stats := []Stats{}
	for rows.Next() {
		s, intIP := Stats{}, int64(0)

		err := rows.Scan(&s.Name, &intIP, &s.Team, &s.Frags, &s.Deaths, &s.Accuracy, &s.Teamkills, &s.Flags, &s.RecordedAt)
		if err != nil {
			log.Println("scanning stats DB row:", err)
		}

		s.IP = ips.Int2IP(intIP).String()
		if s.IP == "255.255.255.255" {
			s.IP = ""
		}

		stats = append(stats, s)
	}

	return stats
}
