package db

import (
	"database/sql"
	"justchess/internal/compression"
	"time"
)

const (
	EasyStockfishId       string = "9W2yEWc2B8yE"
	MediumStockfishId     string = "qaCrtP1_C5qe"
	HardStockfishId       string = "2mHgm-6Jzwt4"
	InsaneStockfishId     string = "dnFoiSQXqwAR"
	ImpossibleStockfishId string = "ORZAsZ5MlE6c"
)

// EngineDifficulty represents the skill level of engine.
type EngineDifficulty int

const (
	Easy EngineDifficulty = iota + 1
	Medium
	Hard
	Insane
	Impossible
)

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
	Checkmate
	Stalemate
	InsufficientMaterial
	FiftyMoves
	ThreefoldRepetition
	Resignation
	Agreement
	TimeForfeit
)

// OptionalPlayer is used to handle possible null players in game tables.
// Guest players are not stored in db even though they can play games.
type OptionalPlayer struct {
	Id     sql.Null[string]
	Name   sql.Null[string]
	Rating sql.Null[float64]
}

type Game struct {
	White       Player             `json:"w"`
	Black       Player             `json:"b"`
	CreatedAt   time.Time          `json:"c"`
	Moves       []compression.Move `json:"m,omitempty"`
	TimeDiffs   []int              `json:"td,omitempty"`
	Id          string             `json:"id"`
	MovesLength int                `json:"ml"`
	Result      Result             `json:"r"`
	Termination Termination        `json:"t"`
	TimeControl int                `json:"tc,omitempty"`
	TimeBonus   int                `json:"tb,omitempty"`
}

// Pagination is used to skip certain amount of game records without using slow
// OFFSET SQL statement. It can be used for all kinds of games.
type Pagination struct {
	CursorCreatedAt time.Time `json:"cca"`
	CursorId        string    `json:"cid"`
}

type GameRepo interface {
	Insert(g Game) error
	// SelectById selects full [Game] information by named id.
	SelectById(id string) (Game, error)
	// SelectByPlayerId selects brief information about [Game]s in which the
	// player with named id took part.
	SelectByPlayerId(id string, p *Pagination) ([]Game, error)
}

type SQLGameRepo struct {
	pool *sql.DB
}

func NewSQLGameRepo(p *sql.DB) SQLGameRepo {
	return SQLGameRepo{pool: p}
}

func (r SQLGameRepo) Insert(g Game) error {
	var wid, bid *string
	if g.White.Id != "" {
		wid = &g.White.Id
	}
	if g.Black.Id != "" {
		wid = &g.Black.Id
	}
	_, err := r.pool.Exec(insertGame, g.Id, wid, bid, g.TimeControl, g.TimeBonus)
	return err
}

func (r SQLGameRepo) SelectById(id string) (Game, error) {
	row := r.pool.QueryRow(selectGameById, id)

	var g Game
	var w, b OptionalPlayer
	var moves, diffs []byte
	if err := row.Scan(
		&g.Id, &moves, &g.MovesLength, &diffs,
		&g.TimeControl, &g.TimeBonus, &g.Result,
		&g.Termination, &g.CreatedAt, &w.Id,
		&w.Name, &w.Rating, &b.Id, &b.Name, &b.Rating,
	); err != nil {
		return Game{}, err
	}

	if w.Id.Valid {
		g.White = Player{Id: w.Id.V, Name: w.Name.V, Rating: w.Rating.V}
	} else {
		g.White = Player{Name: "Guest"}
	}
	if w.Id.Valid {
		g.Black = Player{Id: b.Id.V, Name: b.Name.V, Rating: b.Rating.V}
	} else {
		g.Black = Player{Name: "Guest"}
	}

	if g.Termination != Unterminated {
		g.Moves = compression.HuffmanDecoding(moves, g.MovesLength)
		g.TimeDiffs = compression.DecompressTimeDiffs(diffs, g.MovesLength)
	}
	return g, nil
}

func (r SQLGameRepo) SelectByPlayerId(id string, p *Pagination) ([]Game, error) {
	var rows *sql.Rows
	var err error
	if p == nil {
		rows, err = r.pool.Query(selectGameByPlayerId, id)
	} else {
		rows, err = r.pool.Query(selectGameByPlayerIdPagination, id, p.CursorId, p.CursorCreatedAt)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]Game, 0, 10)
	for rows.Next() {
		var g Game
		var w, b OptionalPlayer
		if err := rows.Scan(
			&g.Id, &g.MovesLength, &g.TimeControl,
			&g.TimeBonus, &g.Result, &g.Termination,
			&g.CreatedAt, &w.Id, &w.Name, &w.Rating,
			&b.Id, &b.Name, &b.Rating,
		); err != nil {
			return nil, err
		}
		if w.Id.Valid {
			g.White = Player{Id: w.Id.V, Name: w.Name.V, Rating: w.Rating.V}
		} else {
			g.White = Player{Name: "Guest"}
		}
		if w.Id.Valid {
			g.Black = Player{Id: b.Id.V, Name: b.Name.V, Rating: b.Rating.V}
		} else {
			g.Black = Player{Name: "Guest"}
		}
		games = append(games, g)
	}
	return games, nil
}

const (
	insertGame = `INSERT INTO
	GAME (ID, WHITE_ID, BLACK_ID, TIME_CONTROL, TIME_BONUS)
VALUES
	($1, $2, $3)`

	selectGameById = `SELECT
	G.ID,
	G.MOVES,
	G.MOVES_LENGTH,
	G.TIME_DIFFERENCES,
	G.TIME_CONTROL,
	G.TIME_BONUS,
	G.RESULT,
	G.TERMINATION,
	G.CREATED_AT,
	W.ID,
	W.NAME,
	W.RATING,
	B.ID,
	B.NAME,
	B.RATING
FROM
	GAME G
	LEFT JOIN PLAYER W ON G.WHITE_ID = W.ID
	LEFT JOIN PLAYER B ON G.BLACK_ID = B.ID
WHERE
	G.ID = $1`

	selectGameByPlayerId = `SELECT
	G.ID,
	G.MOVES_LENGTH,
	G.TIME_CONTROL,
	G.TIME_BONUS,
	G.RESULT,
	G.TERMINATION,
	G.CREATED_AT,
	W.ID,
	W.NAME,
	W.RATING,
	B.ID,
	B.NAME,
	B.RATING
FROM
	GAME G
	LEFT JOIN PLAYER W ON G.WHITE_ID = W.ID
	LEFT JOIN PLAYER B ON G.BLACK_ID = B.ID
WHERE
	(
		G.WHITE_ID = $1
		OR G.BLACK_ID = $1
	)
ORDER BY
	G.CREATED_AT DESC,
	G.ID DESC
LIMIT
	100`

	selectGameByPlayerIdPagination = `SELECT
	G.ID,
	G.MOVES_LENGTH,
	G.TIME_CONTROL,
	G.TIME_BONUS,
	G.RESULT,
	G.TERMINATION,
	G.CREATED_AT,
	W.ID,
	W.NAME,
	W.RATING,
	B.ID,
	B.NAME,
	B.RATING
FROM
	GAME G
	LEFT JOIN PLAYER W ON G.WHITE_ID = W.ID
	LEFT JOIN PLAYER B ON G.BLACK_ID = B.ID
WHERE
	(
		G.WHITE_ID = $1
		OR G.BLACK_ID = $1
	)
	AND (
		(
			G.CREATED_AT = $2
			AND G.ID < $3
		)
		OR G.CREATED_AT < $2
	)
ORDER BY
	G.CREATED_AT DESC,
	G.ID DESC
LIMIT
	100`
)
