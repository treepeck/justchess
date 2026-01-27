package db

import (
	"database/sql"
	"time"
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
		g.termination,
		g.created_at
	FROM game g
	JOIN player w ON g.white_id = w.id
	JOIN player b ON g.black_id = b.id
	WHERE g.id = ? AND g.termination != 1`
)

// Game represents the state of a single completed chess game.
type Game struct {
	CreatedAt   time.Time
	Id          string
	White       Player
	Black       Player
	Control     int
	Bonus       int
	Result      Result
	Termination Termination
}

// Result represents the possible outcomes of a chess game.
type Result int

const (
	Unknown Result = iota
	WhiteWon
	BlackWon
	Draw
)

// Termination represents the reason for the conclusion of the game.  While the
// [Result] types gives the result of the game, it does not provide any extra
// information and so the Termination type is defined for this purpose.
type Termination int

const (
	Unterminated Termination = iota
	Abandoned
	Adjudication
	Normal
	RulesInfraction
	TimeForfeit
)

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
// table.  Error is returned when the game doesn't exist.
func (r GameRepo) SelectById(id string) (Game, error) {
	row := r.pool.QueryRow(selectGameById, id)

	var g Game
	return g, row.Scan(
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
		&g.Termination,
		&g.CreatedAt,
	)
}
