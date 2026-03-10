package db

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/treepeck/chego"
)

// Move represents the completed decoded move.
type Move struct {
	Fen      string `json:"f"`
	San      string `json:"s"`
	TimeDiff int    `json:"t"`
}

// Game represents the state of a single completed chess game.
type Game struct {
	White       Player
	Black       Player
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Moves       json.RawMessage
	Id          string
	MovesLength int
	Control     int
	Bonus       int
	Result      chego.Result
	Termination chego.Termination
}

// ProfileGame represents the short game description to fill up the
// player profile with game history.
type ProfileGame struct {
	CreatedAt   time.Time         `json:"c"`
	Id          string            `json:"i"`
	White       string            `json:"w"`
	Black       string            `json:"b"`
	Result      chego.Result      `json:"r"`
	Termination chego.Termination `json:"t"`
	MovesLength int               `json:"m"`
	Control     int               `json:"ctl"`
	Bonus       int               `json:"bns"`
}

// Pagination is used to skip certain amount of game records without use of slow
// OFFSET SQL statement.
type Pagination struct {
	CursorCreatedAt time.Time `json:"cca"`
	CursorId        string    `json:"cid"`
}

// GameUpdate is used to update the game entity in database.
type GameUpdate struct {
	EncodedMoves    []byte
	CompressedDiffs []byte
	Id              string
	Result          chego.Result
	Termination     chego.Termination
	MovesLength     int
}

// GameRepo provides access to game data.
type GameRepo interface {
	InsertGame(id, whiteId, blackId string, control, bonus int) error
	// SelectById skips abandoned games.
	SelectById(id string) (Game, error)
	// SelectNewestProfileGames selects up to a 100 records of brief data about games
	// in which the player with specified name took part.  The result is ordered by
	// game creation date in descending order, meaning that newer games will go first.
	// Abandoned games are skipped.
	SelectNewestProfileGames(name string) ([]ProfileGame, error)
	// SelectOlderProfileGames is same as [SelectNewestProfileGames] but applies
	// pagination.
	SelectOlderProfileGames(name string, p Pagination) ([]ProfileGame, error)
	UpdateGame(gu GameUpdate) error
	MarkGameAsAbandoned(id string) error
}

// SQLGameRepo wraps the SQL database handle and implements [GameRepo].
type SQLGameRepo struct {
	pool *sql.DB
}

func NewSQLGameRepo(p *sql.DB) SQLGameRepo { return SQLGameRepo{pool: p} }

func (r SQLGameRepo) InsertGame(id, whiteId, blackId string, control, bonus int) error {
	_, err := r.pool.Exec(insertGame, id, whiteId, blackId, control, bonus)
	return err
}

func (r SQLGameRepo) SelectById(id string) (Game, error) {
	row := r.pool.QueryRow(selectGameById, id)

	var g Game
	var encoded []byte
	var compressed []byte
	if err := row.Scan(
		// Scan white player.
		&g.White.Id,
		&g.White.Name,
		&g.White.Rating,
		&g.White.Deviation,
		&g.White.Volatility,
		// Scan black player.
		&g.Black.Id,
		&g.Black.Name,
		&g.Black.Rating,
		&g.Black.Deviation,
		&g.Black.Volatility,
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
		&compressed,
	); err != nil {
		return g, err
	}

	// Decode moves if the game has been terminated.
	if g.Termination != chego.Unterminated {
		moves := make([]Move, g.MovesLength)

		decoded := chego.HuffmanDecoding(encoded, g.MovesLength)
		decompressed := chego.DecompressTimeDiffs(compressed, g.MovesLength)

		for i := range g.MovesLength {
			moves[i] = Move{
				Fen:      decoded[i].Fen,
				San:      decoded[i].San,
				TimeDiff: decompressed[i],
			}
		}

		raw, err := json.Marshal(moves)
		g.Moves = raw
		return g, err
	}
	return g, nil
}

func (r SQLGameRepo) SelectNewestProfileGames(name string) ([]ProfileGame, error) {
	rows, err := r.pool.Query(selectNewestProfileGames, name, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]ProfileGame, 0, 10)
	for rows.Next() {
		var pg ProfileGame
		if err = rows.Scan(
			&pg.White,
			&pg.Black,
			&pg.Result,
			&pg.Termination,
			&pg.Control,
			&pg.Bonus,
			&pg.MovesLength,
			&pg.CreatedAt,
			&pg.Id,
		); err != nil {
			return nil, err
		}
		games = append(games, pg)
	}
	return games, err
}

func (r SQLGameRepo) SelectOlderProfileGames(name string, p Pagination) ([]ProfileGame, error) {
	rows, err := r.pool.Query(
		selectOlderProfileGames,
		name,
		name,
		p.CursorCreatedAt,
		p.CursorId,
		p.CursorCreatedAt,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]ProfileGame, 0, 10)
	for rows.Next() {
		var pg ProfileGame
		if err = rows.Scan(
			&pg.White,
			&pg.Black,
			&pg.Result,
			&pg.Termination,
			&pg.Control,
			&pg.Bonus,
			&pg.MovesLength,
			&pg.CreatedAt,
			&pg.Id,
		); err != nil {
			log.Print(err)
			return nil, err
		}
		games = append(games, pg)
	}
	return games, err
}

func (r SQLGameRepo) UpdateGame(gu GameUpdate) error {
	_, err := r.pool.Exec(
		updateGame,
		gu.Result, gu.Termination,
		gu.MovesLength, gu.EncodedMoves,
		gu.CompressedDiffs, gu.Id,
	)
	return err
}

func (r SQLGameRepo) MarkGameAsAbandoned(id string) error {
	_, err := r.pool.Exec(markGameAsAbandoned, id)
	return err
}

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
		g.updated_at,
		g.time_differences
	FROM game g
	INNER JOIN player w ON g.white_id = w.id
	INNER JOIN player b ON g.black_id = b.id
	WHERE g.id = ? AND g.termination != 1`

	selectNewestProfileGames = `
	SELECT
		w.name AS w_name,
		b.name AS b_name,
	    g.result,
	    g.termination,
	    g.time_control,
   	    g.time_bonus,
	    g.moves_length,
	    g.created_at,
	    g.id
	FROM game g
	INNER JOIN player w ON g.white_id = w.id
	INNER JOIN player b ON g.black_id = b.id
	WHERE
		(w.name = ? OR b.name = ?)
	    AND g.termination != 1
	ORDER BY g.created_at DESC, g.id DESC
	LIMIT 100`

	// Same as [selectNewestProfileGames] by with cursor-based pagination.
	// Select 100 games which are older than provided timestamp.
	selectOlderProfileGames = `
	SELECT
		w.name AS w_name,
		b.name AS b_name,
		g.result,
		g.termination,
		g.time_control,
		g.time_bonus,
		g.moves_length,
		g.created_at,
		g.id
	FROM game g
	INNER JOIN player w ON g.white_id = w.id
	INNER JOIN player b ON g.black_id = b.id
	WHERE
		(w.name = ? OR b.name = ?)
		AND g.termination != 1
		AND (
			(g.created_at = ? AND g.id < ?)
	        OR g.created_at < ?
	    )
	ORDER BY g.created_at DESC, g.id DESC
	LIMIT 100`

	// Updates the result, termination, moves, and updated_at columns of a single
	// game record.
	updateGame = `
	UPDATE game
	SET
		result = ?,
		termination = ?,
		moves_length = ?,
		moves = ?,
		time_differences = ?,
		updated_at = CURRENT_TIMESTAMP
	WHERE game.id = ?`

	markGameAsAbandoned = `	UPDATE game SET termination = 1	WHERE game.id = ?`
)
