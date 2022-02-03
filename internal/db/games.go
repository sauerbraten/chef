package db

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/sauerbraten/chef/pkg/extinfo"
)

type Game struct {
	Server
	ID             int64              `json:"id"`
	MasterMode     extinfo.MasterMode `json:"master_mode"`
	GameMode       extinfo.GameMode   `json:"game_mode"`
	Map            string             `json:"map"`
	StartedAt      int64              `json:"started_at"`
	SecondsLeft    int64              `json:"seconds_left"`
	LastRecordedAt int64              `json:"last_recorded_at"`
	EndedAt        *int64             `json:"ended_at"`
}

// GetGameID returns the ID of a game, matched using game mode, map and start time.
// In case no matching game exists, a new one is inserted and the rowid of the new entry is returned.
func (db *Database) GetGameID(mastermode, gamemode int8, mapname string, serverID int64, secsLeft, scanInterval int) (gameID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	secsPlayed := 10*60 - secsLeft // assumes 10 minute games
	startedAt := time.Now().Add(-time.Duration(secsPlayed) * time.Second)
	startedAfter := startedAt.Add(-5 * time.Minute) // 5 minutes buffer for pauses/overtimes

	err := db.QueryRow(
		"select id from games where server = ? and game_mode = ? and map = ? and started_at > ? and (seconds_left >= ? or (seconds_left <= ? and ? <= 120))",
		serverID, gamemode, mapname, startedAfter.Unix(), secsLeft, scanInterval, secsLeft,
	).Scan(&gameID)

	if err == sql.ErrNoRows {
		prevGameEndedAt := time.Now().Unix() - int64(secsLeft)
		_, err := db.Exec(
			"update games set ended_at = ?, seconds_left = max(0, seconds_left - (?-last_recorded_at)) where ended_at is null and server = ?",
			prevGameEndedAt, prevGameEndedAt, serverID,
		)
		if err != nil {
			log.Fatalln("marking past game(s) on server as ended:", err)
		}

		res, err := db.Exec(
			"insert into games (server, master_mode, game_mode, map, started_at, seconds_left) values (?, ?, ?, ?, ?, ?)",
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
		_, err = db.Exec("update games set master_mode = ?, seconds_left = ? where id = ?", mastermode, secsLeft, gameID)
		if err != nil {
			log.Fatalln("updating 'seconds_left' in game:", err)
		}
	}

	return
}

func (db *Database) UpdateGameLastRecordedAt(gameID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, err := db.Exec("update games set last_recorded_at = strftime('%s', 'now') where id = ?", gameID)
	if err != nil {
		log.Fatalln("updating game's 'last recorded at' time:", err)
	}
}

func (db *Database) SetGameEnded(gameID int64) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, err := db.Exec("update games set ended_at = strftime('%s', 'now') where id = ?", gameID)
	if err != nil {
		log.Fatalln("error updating game's 'ended at' time:", err)
	}
}

type GamesLookupQuery struct {
	PlayerName string           `json:"player_name,omitempty"`
	Map        string           `json:"map,omitempty"`
	GameMode   extinfo.GameMode `json:"game_mode"`
}

type FinishedGamesLookup struct {
	Query   GamesLookupQuery `json:"query"`
	Results []Game           `json:"results"`
}

func (db *Database) LookupGame(q GamesLookupQuery) FinishedGamesLookup {
	conditions := []string{}
	args := []interface{}{}

	if q.PlayerName != "" {
		conditions = append(conditions, "games.id in (select game from stats where combination in (select id from combinations where name in (select id from names where name like ?)))")
		args = append(args, "%"+q.PlayerName+"%")
	}

	if q.Map != "" {
		conditions = append(conditions, "map like ?")
		args = append(args, q.Map)
	}

	if q.GameMode >= 0 {
		conditions = append(conditions, "game_mode = ?")
		args = append(args, q.GameMode)
	}

	const (
		columns      = "games.id, game_mode, map, started_at, ended_at, server, servers.ip, servers.port, servers.description, servers.mod"
		joinedTables = "games, servers on games.server = servers.id"
		sorting      = "started_at desc"
	)

	query := "select " + columns + " from " + joinedTables + " where " + strings.Join(conditions, " and ") + " order by " + sorting + " limit 1000"

	db.mutex.Lock()
	defer db.mutex.Unlock()

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Fatalln("performing game look up:", err)
	}
	defer rows.Close()

	games := []Game{}

	for rows.Next() {
		game := Game{}
		srv := &game.Server

		err := rows.Scan(&game.ID, &game.GameMode, &game.Map, &game.StartedAt, &game.EndedAt, &srv.ID, &srv.IP, &srv.Port, &srv.Description, &srv.Mod)
		if err != nil {
			log.Println("scanning DB row into game:", err)
		}

		games = append(games, game)
	}

	return FinishedGamesLookup{
		Query:   q,
		Results: games,
	}
}

func (db *Database) GetGame(id int64) *Game {
	const (
		columns      = "games.id, master_mode, game_mode, map, started_at, ended_at, server, servers.ip, servers.port, servers.description, servers.mod"
		joinedTables = "games, servers on games.server = servers.id"
		condition    = "games.id = ?"
	)

	query := "select " + columns + " from " + joinedTables + " where " + condition

	db.mutex.Lock()
	defer db.mutex.Unlock()

	game := &Game{}
	srv := &game.Server

	err := db.QueryRow(query, id).Scan(&game.ID, &game.MasterMode, &game.GameMode, &game.Map, &game.StartedAt, &game.EndedAt, &srv.ID, &srv.IP, &srv.Port, &srv.Description, &srv.Mod)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Fatalf("fetching game from DB: %s: %v\n", query, err)
	}

	return game
}
