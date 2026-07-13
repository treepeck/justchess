package db

import (
	"database/sql"
	"time"

	"github.com/treepeck/chego"
)

// OptionalPlayer is used to handle possible null players in game tables.
// Guest players are not stored in db even though they can play against engine.
type OptionalPlayer struct {
	Id     sql.Null[string]
	Name   sql.Null[string]
	Rating sql.Null[float64]
}

// GameBrief is a brief information about a game of arbitrary kind.
type GameBrief struct {
	CreatedAt   time.Time         `json:"ca"`
	Id          string            `json:"i"`
	Termination chego.Termination `json:"t"`
	Result      chego.Result      `json:"r"`
	MovesLength int               `json:"ml"`
}

// RatedBrief is used to fill up the rated games section of the game table
// on player profile page.
type RatedBrief struct {
	GameBrief
	White       Player `json:"w"`
	Black       Player `json:"b"`
	TimeControl int    `json:"tc"`
	TimeBonus   int    `json:"tb"`
}

// EngineBrief is used to fill up the engine games section of the game table
// on player profile page.
type EngineBrief struct {
	GameBrief
	Player      Player           `json:"p"`
	PlayerColor int              `json:"pc"`
	Difficulty  EngineDifficulty `json:"d"`
}

// EngineDifficulty represents the skill level of engine.
type EngineDifficulty int

const (
	Easy EngineDifficulty = iota + 1
	Medium
	Hard
	Insane
	Impossible
)

// Pagination is used to skip certain amount of game records without using slow
// OFFSET SQL statement. It can be used for all kinds of games.
type Pagination struct {
	CursorCreatedAt time.Time `json:"cca"`
	CursorId        string    `json:"cid"`
}

type GameRepo interface {
	SelectRatedBrief(playerId string, p *Pagination) ([]RatedBrief, error)
	InsertEngineGame(id, playerId string, c chego.Color, d EngineDifficulty) error
	SelectEngineBrief(playerId string, p *Pagination) ([]EngineBrief, error)
}

type SQLGameRepo struct {
	pool *sql.DB
}

func NewSQLGameRepo(p *sql.DB) SQLGameRepo { return SQLGameRepo{pool: p} }

func (r SQLGameRepo) SelectRatedBrief(playerId string, p *Pagination) ([]RatedBrief, error) {
	var rows *sql.Rows
	var err error
	if p == nil {
		rows, err = r.pool.Query(selectRatedBrief, playerId)
	} else {
		rows, err = r.pool.Query(selectPaginatedRatedBrief, playerId, p.CursorCreatedAt, p.CursorId)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]RatedBrief, 0, 10)
	for rows.Next() {
		var r RatedBrief
		var w OptionalPlayer
		var b OptionalPlayer
		if err = rows.Scan(
			&r.Id, &r.Termination, &r.Result, &r.MovesLength, &r.CreatedAt,
			&r.TimeControl, &r.TimeBonus, &w.Id, &w.Name, &w.Rating,
			&b.Id, &b.Name, &b.Rating,
		); err != nil {
			return nil, err
		}
		if w.Id.Valid {
			r.White = Player{Id: w.Id.V, Name: w.Name.V, Rating: w.Rating.V}
		} else {
			r.White = Player{Name: "Guest"}
		}
		if b.Id.Valid {
			r.Black = Player{Id: b.Id.V, Name: b.Name.V, Rating: b.Rating.V}
		} else {
			r.Black = Player{Name: "Guest"}
		}
		games = append(games, r)
	}
	return games, err
}

func (r SQLGameRepo) InsertEngineGame(id, playerId string, c chego.Color, d EngineDifficulty) error {
	_, err := r.pool.Exec(insertGame, id)
	if err != nil {
		return nil
	}
	if playerId == "" { // Handle Guest player.
		_, err = r.pool.Exec(insertEngineGame, id, nil, c, d)
	} else {
		_, err = r.pool.Exec(insertEngineGame, id, playerId, c, d)
	}
	return nil
}

func (r SQLGameRepo) SelectEngineBrief(playerId string, p *Pagination) ([]EngineBrief, error) {
	var rows *sql.Rows
	var err error
	if p == nil {
		rows, err = r.pool.Query(selectEngineBrief, playerId)
	} else {
		rows, err = r.pool.Query(selectPaginatedEngineBrief, playerId, p.CursorCreatedAt, p.CursorId)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]EngineBrief, 0, 10)
	for rows.Next() {
		var e EngineBrief
		var p OptionalPlayer
		if err = rows.Scan(
			&e.Id, &e.Termination, &e.Result, &e.MovesLength, &e.CreatedAt,
			&e.PlayerColor, &e.Difficulty, &p.Id, &p.Name, &p.Rating,
		); err != nil {
			return nil, err
		}
		if p.Id.Valid {
			e.Player = Player{Id: p.Id.V, Name: p.Name.V, Rating: p.Rating.V}
		} else {
			e.Player = Player{Name: "Guest"}
		}
		games = append(games, e)
	}
	return games, err
}

const (
	insertGame = `INSERT INTO game (id) VALUES ($1)`

	selectRatedBrief = `SELECT g.id, g.termination, g.result, g.moves_length, g.created_at,
		r.time_control, r.time_bonus,
		w.id, w.name, w.rating, b.id, b.name, b.rating
	FROM game g
	INNER JOIN rated_game r ON r.game_id = g.id
	INNER JOIN player w ON r.white_id = w.id
	INNER JOIN player b ON r.black_id = b.id
	WHERE (r.white_id = $1 OR r.black_id = $1)
	ORDER BY g.created_at DESC, g.id DESC
	LIMIT 100`

	selectPaginatedRatedBrief = `SELECT g.id, g.termination, g.result, g.moves_length, g.created_at,
		r.time_control, r.time_bonus,
		w.id, w.name, w.rating, b.id, b.name, b.rating
	FROM game g
	INNER JOIN rated_game r ON r.game_id = g.id
	INNER JOIN player w ON r.white_id = w.id
	INNER JOIN player b ON r.black_id = b.id
	WHERE (r.white_id = $1 OR r.black_id = $1)
		AND (
			(g.created_at = $2 AND g.id < $3)
			OR g.created_at < $2
		)
	ORDER BY g.created_at DESC, g.id DESC
	LIMIT 100`

	insertEngineGame = `INSERT INTO engine_game (game_id, player_id, player_color, difficulty)
	VALUES ($1, $2, $3, $4)`

	selectEngineBrief = `SELECT g.id, g.termination, g.result, g.moves_length, g.created_at,
		e.player_color, e.difficulty,
		p.id, p.name, p.rating
	FROM game g
	INNER JOIN engine_game e ON e.game_id = g.id
	INNER JOIN player p ON e.player_id = p.id
	WHERE e.player_id = $1
	ORDER BY g.created_at DESC, g.id DESC
	LIMIT 100`

	selectPaginatedEngineBrief = `SELECT g.id, g.termination, g.result, g.moves_length, g.created_at,
		e.player_color, e.difficulty,
		p.id, p.name, p.rating
	FROM game g
	INNER JOIN engine_game e ON e.game_id = g.id
	INNER JOIN player p ON e.player_id = p.id
	WHERE e.player_id = $1
		AND (
			(g.created_at = $2 AND g.id < $3)
			OR g.created_at < $2
		)
	ORDER BY g.created_at DESC, g.id DESC
	LIMIT 100`
)
