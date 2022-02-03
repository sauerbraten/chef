package db

import (
	"database/sql"
	"log"
	"time"
)

type Game struct {
	ID int64 `json:"id"`
	Server
	MasterMode     int    `json:"master_mode"`
	GameMode       int    `json:"game_mode"`
	Map            string `json:"map"`
	StartedAt      int64  `json:"started_at"`
	SecondsLeft    int64  `json:"seconds_left"`
	LastRecordedAt int64  `json:"last_recorded_at"`
	EndedAt        int64  `json:"ended_at"`
}

// GetGameID returns the ID of a game, matched using game mode, map and start time.
// In case no matching game exists, a new one is inserted and the rowid of the new entry is returned.
func (db *Database) GetGameID(mastermode, gamemode int8, mapname string, serverID int64, secsLeft, scanInterval int) (gameID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	secsPlayed := 10*60 - secsLeft // assumes 10 minute rounds
	startedAt := time.Now().Add(-time.Duration(secsPlayed) * time.Second)
	startedAfter := startedAt.Add(-5 * time.Minute) // 5 minutes buffer for pauses/overtimes

	err := db.QueryRow(
		"select `id` from `games` where `server` = ? and `game_mode` = ? and `map` = ? and `started_at` > ? and (`seconds_left` >= ? or (`seconds_left` <= ? and ? <= 120))",
		serverID, gamemode, mapname, startedAfter.Unix(), secsLeft, scanInterval, secsLeft,
	).Scan(&gameID)

	if err == sql.ErrNoRows {
		prevGameEndedAt := time.Now().Unix() - int64(secsLeft)
		_, err := db.Exec(
			"update `games` set `ended_at` = ?, `seconds_left` = max(0, `seconds_left` - (?-`last_recorded_at`)) where `ended_at` is null and `server` = ?",
			prevGameEndedAt, prevGameEndedAt, serverID,
		)
		if err != nil {
			log.Fatalln("marking past game(s) on server as ended:", err)
		}

		res, err := db.Exec(
			"insert into `games` (`server`, `master_mode`, `game_mode`, `map`, `started_at`, `seconds_left`) values (?, ?, ?, ?, ?, ?)",
			serverID, mastermode, gamemode, mapname, startedAt.Unix(), secsLeft,
		)
		if err != nil {
			log.Fatalln("inserting new game into database:", err)
		}

		gameID, err = res.LastInsertId()
		if err != nil {
			log.Fatalln("getting ID of newly inserted game:", err)
		}
	} else if err != nil {
		log.Fatalln("getting ID of game:", err)
	} else {
		_, err = db.Exec("update `games` set `master_mode` = ?, `seconds_left` = ? where `id` = ?", mastermode, secsLeft, gameID)
		if err != nil {
			log.Fatalln("updating 'seconds_left' in game:", err)
		}
	}

	return
}

func (db *Database) UpdateGameLastRecordedAt(gameID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, err := db.Exec("update `games` set `last_recorded_at` = strftime('%s', 'now') where `id` = ?", gameID)
	if err != nil {
		log.Fatalln("updating game's 'last recorded at' time:", err)
	}
}

func (db *Database) SetGameEnded(gameID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, err := db.Exec("update `games` set `ended_at` = strftime('%s', 'now') where `id` = ?", gameID)
	if err != nil {
		log.Fatalln("error updating game's 'ended at' time:", err)
	}
}
