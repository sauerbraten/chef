package db

import (
	"log"
)

type Score struct {
	Team   string `json:"team"`
	Points int    `json:"points"`
}

func (db *Database) UpdateScore(gameID int64, team string, points int) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, err := db.Exec(`insert or replace into "scores" ("game", "team", "points") values (?, ?, ?)`, gameID, team, points)
	if err != nil {
		log.Fatalf("error updating game's team score (game %d, team %s, new score %d): %v\n", gameID, team, points, err)
	}
}

func (db *Database) GetScores(gameID int64) []Score {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	rows, err := db.Query(`select "team", "points" from "scores" where "game" = ?`, gameID)
	if err != nil {
		log.Fatalf("fetching scores from DB: %v\n", err)
	}

	scores := []Score{}
	for rows.Next() {
		s := Score{}

		err := rows.Scan(&s.Team, &s.Points)
		if err != nil {
			log.Println("scanning scores DB row:", err)
		}

		scores = append(scores, s)
	}

	return scores
}
