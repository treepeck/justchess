package db

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/treepeck/chego"
)

// The following block of constants declares SQL queries used to access and
// modify the game table.
const (
	insertGame = `
	INSERT INTO game (
		id,
		white_id,
		black_id,
		time_control,
		time_bonus
	)
	VALUES (?, ?, ?, ?, ?)`

	// Select game by id excluding the abandoned games.
	selectGameById = `
	SELECT
		w.id AS w_id,
		w.name AS w_name,
		w.rating AS w_rating,
		w.rating_deviation AS w_rating_deviation,
		w.rating_volatility AS w_rating_volatility,

		b.id AS b_id,
		b.name AS b_name,
		b.rating AS b_rating,
		b.rating_deviation AS b_rating_deviation,
		b.rating_volatility AS b_rating_volatility,

		g.id,
		g.time_control,
		g.time_bonus,
		g.result,
		g.moves_length,
		g.moves,
		g.termination,
		g.created_at,
		g.updated_at
	FROM game g
	JOIN player w ON g.white_id = w.id
	JOIN player b ON g.black_id = b.id
	WHERE g.id = ? AND g.termination != 1`

	// Updates the result, termination, moves, and updated_at columns of a single
	// game record.
	updateGame = `
	UPDATE game
	SET
		result = ?,
		termination = ?,
		moves_length = ?,
		moves = ?,
		updated_at = CURRENT_TIMESTAMP
	WHERE game.id = ?
	`
)

// Game represents the state of a single completed chess game.
type Game struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Id          string
	MovesLength int
	Moves       json.RawMessage
	White       Player
	Black       Player
	Control     int
	Bonus       int
	Result      chego.Result
	Termination chego.Termination
}

// GameRepo wraps the database connection pool and provides methods to access
// and modify the game table.
type GameRepo struct {
	pool *sql.DB
}

func NewGameRepo(p *sql.DB) GameRepo { return GameRepo{pool: p} }

// Insert inserts a single record into the game table.
func (r GameRepo) Insert(id, whiteId, blackId string, control, bonus int) error {
	_, err := r.pool.Exec(insertGame, id, whiteId, blackId, control, bonus)
	return err
}

// SelectById selects a single record with the specified id from the game
// table.  Error is returned when the game doesn't exist or was abandoned.
//
// If the game has been terminated, the moves will be decoded.
// Otherwise, an empty array will be returned.
func (r GameRepo) SelectById(id string) (Game, error) {
	row := r.pool.QueryRow(selectGameById, id)

	var g Game
	var encoded []byte
	if err := row.Scan(
		// Scan white player.
		&g.White.Id,
		&g.White.Name,
		&g.White.Rating,
		&g.White.RatingDeviation,
		&g.White.RatingVolatility,
		// Scan black player.
		&g.Black.Id,
		&g.Black.Name,
		&g.Black.Rating,
		&g.Black.RatingDeviation,
		&g.Black.RatingVolatility,
		// Scan game data.
		&g.Id,
		&g.Control,
		&g.Bonus,
		&g.Result,
		&g.MovesLength,
		&encoded,
		&g.Termination,
		&g.CreatedAt,
		&g.UpdatedAt,
	); err != nil {
		return g, err
	}

	// Decode moves if the game has been terminated.
	if g.Termination != chego.Unterminated {
		raw, err := json.Marshal(chego.HuffmanDecoding(encoded, g.MovesLength))
		if err == nil {
			g.Moves = raw
		}
	}
	return g, nil
}

// Update updates a single record with the specified id in the game table by
// setting the table columns to the provided values.
func (r GameRepo) Update(res chego.Result, t chego.Termination, movesLength int, moves []byte, id string) error {
	_, err := r.pool.Exec(updateGame, res, t, movesLength, moves, id)
	return err
}
