package db

import (
	"log"
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
		"insert into `stats` (`combination`, `game`, `team`, `frags`, `deaths`, `accuracy`, `teamkills`, `flags`) values (?, ?, ?, ?, ?, ?, ?, ?)",
		combinationID, gameID, team, frags, deaths, accuracy, teamkills, flags)
	if err != nil {
		log.Fatalf("error inserting stats (comb %v, game %v): %v\n", combinationID, gameID, err)
	}
}
