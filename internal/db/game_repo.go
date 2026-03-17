package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/treepeck/chego"
)

// RatedGame represents the state of a single rated game.
type RatedGame struct {
	White       Player
	Black       Player
	Moves       []chego.PlayedMove
	TimeDiffs   []int
	Id          string
	MovesLength int
	Control     int
	Bonus       int
	Result      chego.Result
	Termination chego.Termination
}

// RatedGameBrief represents a brief rated game description to fill up
// the player profile page with game history.
type RatedGameBrief struct {
	CreatedAt   time.Time         `json:"c"`
	Id          string            `json:"i"`
	WhiteName   string            `json:"w"`
	BlackName   string            `json:"b"`
	Result      chego.Result      `json:"r"`
	Termination chego.Termination `json:"t"`
	MovesLength int               `json:"m"`
	Control     int               `json:"ctl"`
	Bonus       int               `json:"bns"`
}

// RatedGameUpdate is used to update the rated game entity in database.
type RatedGameUpdate struct {
	EncodedMoves    []byte
	CompressedDiffs []byte
	Id              string
	Result          chego.Result
	Termination     chego.Termination
	MovesLength     int
}

// EngineGame represents the state of a single game played vs engine.
type EngineGame struct {
	Player      Player
	Moves       []chego.PlayedMove
	Id          string
	PlayerColor chego.Color
	Result      chego.Result
	Termination chego.Termination
	MovesLength int
}

// EngineGameBrief represents a brief engine game description to fill up
// the player profile page with game history.
type EngineGameBrief struct {
	CreatedAt   time.Time         `json:"c"`
	Id          string            `json:"i"`
	Result      chego.Result      `json:"r"`
	Termination chego.Termination `json:"t"`
	MovesLength int               `json:"m"`
}

// RatedGameUpdate is used to update the engine game entity in database.
type EngineGameUpdate struct {
	EncodedMoves []byte
	Id           string
	Result       chego.Result
	Termination  chego.Termination
	MovesLength  int
}

// Pagination is used to skip certain amount of game records without use of slow
// OFFSET SQL statement. Can be used for all kinds of games.
type Pagination struct {
	CursorCreatedAt time.Time `json:"cca"`
	CursorId        string    `json:"cid"`
}

// GameRepo provides access to game data.
// SelectOlder* is same as SelectNewest* but applies pagination.
type GameRepo interface {
	InsertRated(id, whiteId, blackId string, control, bonus int) error
	SelectRated(id string) (RatedGame, error)
	SelectNewestRated(id string) ([]RatedGameBrief, error)
	SelectOlderRated(id string, p Pagination) ([]RatedGameBrief, error)
	UpdateRated(gu RatedGameUpdate) error
	MarkRatedAsAbandoned(id string) error

	InsertEngine(id, playerId string, playerColor chego.Color) error
	SelectEngine(id string) (EngineGame, error)
	SelectNewestEngine(id string) ([]EngineGameBrief, error)
	SelectOlderEngine(id string, p Pagination) ([]EngineGameBrief, error)
	UpdateEngine(gu EngineGameUpdate) error
	MarkEngineAsAbandoned(id string) error
}

// SQLGameRepo wraps the SQL database handle and implements [GameRepo].
type SQLGameRepo struct {
	pool *sql.DB
}

func NewSQLGameRepo(p *sql.DB) SQLGameRepo { return SQLGameRepo{pool: p} }

func (r SQLGameRepo) InsertRated(id, whiteId, blackId string, control, bonus int) error {
	_, err := r.pool.Exec(insertRated, id, whiteId, blackId, control, bonus)
	return err
}

func (r SQLGameRepo) SelectRated(id string) (RatedGame, error) {
	row := r.pool.QueryRow(selectRated, id)

	var g RatedGame
	var encoded, compressed []byte
	if err := row.Scan(
		// Scan white player.
		&g.White.Id, &g.White.Name, &g.White.Rating,
		&g.White.Deviation, &g.White.Volatility,
		// Scan black player.
		&g.Black.Id, &g.Black.Name, &g.Black.Rating,
		&g.Black.Deviation, &g.Black.Volatility,
		// Scan game data.
		&g.Id, &g.Control, &g.Bonus, &g.Result, &g.MovesLength,
		&encoded, &g.Termination, &compressed,
	); err != nil {
		return g, err
	}

	// Decode moves if the game has been terminated.
	if g.Termination != chego.Unterminated {
		g.Moves = chego.HuffmanDecoding(encoded, g.MovesLength)
		g.TimeDiffs = chego.DecompressTimeDiffs(compressed, g.MovesLength)
	}
	return g, nil
}

func (r SQLGameRepo) SelectNewestRated(id string) ([]RatedGameBrief, error) {
	rows, err := r.pool.Query(selectNewestRated, id, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]RatedGameBrief, 0, 10)
	for rows.Next() {
		var pg RatedGameBrief
		if err = rows.Scan(
			&pg.WhiteName, &pg.BlackName, &pg.Result, &pg.Termination,
			&pg.Control, &pg.Bonus, &pg.MovesLength, &pg.CreatedAt, &pg.Id,
		); err != nil {
			return nil, err
		}
		games = append(games, pg)
	}
	return games, err
}

func (r SQLGameRepo) SelectOlderRated(id string, p Pagination) ([]RatedGameBrief, error) {
	rows, err := r.pool.Query(
		selectOlderRated, id, id, p.CursorCreatedAt,
		p.CursorId, p.CursorCreatedAt,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]RatedGameBrief, 0, 10)
	for rows.Next() {
		var pg RatedGameBrief
		if err = rows.Scan(
			&pg.WhiteName, &pg.BlackName, &pg.Result, &pg.Termination,
			&pg.Control, &pg.Bonus, &pg.MovesLength, &pg.CreatedAt, &pg.Id,
		); err != nil {
			log.Print(err)
			return nil, err
		}
		games = append(games, pg)
	}
	return games, err
}

func (r SQLGameRepo) UpdateRated(gu RatedGameUpdate) error {
	_, err := r.pool.Exec(
		updateRated, gu.Result, gu.Termination, gu.MovesLength,
		gu.EncodedMoves, gu.CompressedDiffs, gu.Id,
	)
	return err
}

func (r SQLGameRepo) MarkRatedAsAbandoned(id string) error {
	_, err := r.pool.Exec(markRatedAsAbandoned, id)
	return err
}

func (r SQLGameRepo) InsertEngine(id, playerId string, c chego.Color) error {
	_, err := r.pool.Exec(insertEngine, id, playerId, c)
	return err
}

func (r SQLGameRepo) SelectEngine(id string) (EngineGame, error) {
	row := r.pool.QueryRow(selectEngine, id)

	var g EngineGame
	var encoded []byte
	err := row.Scan(
		&g.Player.Id, &g.Player.Name, &g.Player.Rating,
		&g.Player.Deviation, &g.Player.Volatility,
		&g.Id, &g.Result, &g.Termination, &g.MovesLength,
		&encoded, &g.PlayerColor,
	)
	if err != nil {
		return g, err
	}

	// Decode moves if the game has been terminated.
	if g.Termination != chego.Unterminated {
		g.Moves = chego.HuffmanDecoding(encoded, g.MovesLength)
	}
	return g, nil
}

func (r SQLGameRepo) SelectNewestEngine(id string) ([]EngineGameBrief, error) {
	rows, err := r.pool.Query(selectNewestEngine, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]EngineGameBrief, 0, 10)
	for rows.Next() {
		var pg EngineGameBrief
		if err = rows.Scan(); err != nil {
			return nil, err
		}
		games = append(games, pg)
	}
	return games, err
}

func (r SQLGameRepo) SelectOlderEngine(id string, p Pagination) ([]EngineGameBrief, error) {
	panic("not implemented")
}

func (r SQLGameRepo) UpdateEngine(gu EngineGameUpdate) error {
	_, err := r.pool.Exec(
		updateEngine, gu.Result, gu.Termination,
		gu.MovesLength, gu.EncodedMoves, gu.Id,
	)
	return err
}

func (r SQLGameRepo) MarkEngineAsAbandoned(id string) error {
	_, err := r.pool.Exec(markEngineAsAbandoned, id)
	return err
}

const (
	insertRated = `
	INSERT INTO rated_game (
		id,
		white_id,
		black_id,
		time_control,
		time_bonus
	)
	VALUES (?, ?, ?, ?, ?)`

	selectRated = `
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
		g.time_differences
	FROM rated_game g
	INNER JOIN player w ON g.white_id = w.id
	INNER JOIN player b ON g.black_id = b.id
	WHERE g.id = ? AND g.termination != 1`

	selectNewestRated = `
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
	FROM rated_game g
	INNER JOIN player w ON g.white_id = w.id
	INNER JOIN player b ON g.black_id = b.id
	WHERE
		(g.white_id = ? OR g.black_id = ?)
	    AND g.termination != 1
	ORDER BY g.created_at DESC, g.id DESC
	LIMIT 100`

	selectOlderRated = `
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
	FROM rated_game g
	INNER JOIN player w ON g.white_id = w.id
	INNER JOIN player b ON g.black_id = b.id
	WHERE
		(g.white_id = ? OR g.black_id = ?)
		AND g.termination != 1
		AND (
			(g.created_at = ? AND g.id < ?)
	        OR g.created_at < ?
	    )
	ORDER BY g.created_at DESC, g.id DESC
	LIMIT 100`

	updateRated = `
	UPDATE rated_game
	SET
		result = ?,
		termination = ?,
		moves_length = ?,
		moves = ?,
		time_differences = ?,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = ?`

	markRatedAsAbandoned = `UPDATE rated_game SET termination = 1 WHERE id = ?`

	insertEngine = `
	INSERT INTO engine_game (
		id,
		player_id,
		player_color
	)
	VALUES (?, ?, ?)`

	selectEngine = `
	SELECT
		p.id AS p_id,
		p.name AS p_name,
		p.rating AS p_rating,
		p.rating_deviation AS p_rating_deviation,
		p.rating_volatility AS p_rating_volatility,

		g.id,
		g.result,
		g.termination,
		g.moves_length,
		g.moves,
		g.player_color
	FROM engine_game g
	INNER JOIN player p ON g.player_id = p.id
	WHERE g.id = ? AND g.termination != 1`

	selectNewestEngine = `
	SELECT
	    result,
	    termination,
	    moves_length,
	    created_at,
	    id
	FROM engine_game
	WHERE
		player_id = ? AND termination != 1
	ORDER BY created_at DESC, id DESC
	LIMIT 100`

	selectOlderEngine = `
	SELECT
		result,
	    termination,
	    moves_length,
	    created_at,
	    id
	FROM engine_game
	WHERE
		(player_id = ? AND termination != 1)
		AND (
			(created_at = ? AND g.id < ?)
	        OR created_at < ?
	    )
	ORDER BY created_at DESC, id DESC
	LIMIT 100`

	updateEngine = `
	UPDATE engine_game
	SET
		result = ?,
		termination = ?,
		moves_length = ?,
		moves = ?,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = ?`

	markEngineAsAbandoned = `UPDATE engine_game SET termination = 1 WHERE id = ?`
)
