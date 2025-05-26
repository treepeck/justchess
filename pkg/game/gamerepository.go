package game

import (
	"errors"
	"justchess/pkg/chess"
	"justchess/pkg/chess/bitboard"
	"justchess/pkg/chess/enums"
	"justchess/pkg/chess/fen"
	"justchess/pkg/db"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type GameDTO struct {
	Id          uuid.UUID             `json:"id"`
	WhiteId     uuid.UUID             `json:"whiteId"`
	BlackId     uuid.UUID             `json:"blackId"`
	TimeControl int                   `json:"timeControl"`
	TimeBonus   int                   `json:"timeBonus"`
	Result      enums.Result          `json:"result"`
	Winner      enums.Color           `json:"winner"`
	Moves       []chess.CompletedMove `json:"moves"`
	CreatedAt   time.Time             `json:"createdAt"`
	WhiteName   string                `json:"whiteName"`
	BlackName   string                `json:"blackName"`
}

// shortGameDTO represents a completed game without completed moves.
type shortGameDTO struct {
	Id          uuid.UUID    `json:"id"`
	WhiteId     uuid.UUID    `json:"wid"`
	BlackId     uuid.UUID    `json:"bid"`
	Result      enums.Result `json:"r"`
	Winner      enums.Color  `json:"w"`
	MovesLen    int          `json:"m"`
	WhiteName   string       `json:"wn"`
	BlackName   string       `json:"bn"`
	TimeControl int          `json:"tc"`
	TimeBonus   int          `json:"tb"`
	CreatedAt   time.Time    `json:"ca"`
}

func selectById(id string) (g GameDTO, err error) {
	query :=
		`SELECT
			game.*,
			white_player.name AS white_name,
			black_player.name AS black_name
		FROM game
		JOIN player AS white_player ON game.white_id = white_player.id
		JOIN player AS black_player ON game.black_id = black_player.id 
		WHERE game.id = $1;`

	rows, err := db.Pool.Query(query, id)
	if err != nil {
		return
	}

	if !rows.Next() {
		return g, errors.New("game not found")
	}
	var compressedMoves []int32
	err = rows.Scan(&g.Id, &g.WhiteId, &g.BlackId, &g.TimeControl, &g.TimeBonus,
		&g.Result, &g.Winner, pq.Array(&compressedMoves), &g.CreatedAt,
		&g.WhiteName, &g.BlackName)
	g.Moves = decompressMoves(compressedMoves, fen.DefaultFEN, g.TimeControl)
	return
}

func selectByPlayerId(id string) (games []shortGameDTO, err error) {
	query :=
		`SELECT
			game.*,
			white_player.name AS white_name,
			black_player.name AS black_name
		FROM game
		JOIN player AS white_player ON game.white_id = white_player.id
		JOIN player AS black_player ON game.black_id = black_player.id
		WHERE game.white_id = $1 OR game.black_id = $1
		ORDER BY game.created_at DESC LIMIT 100;`

	rows, err := db.Pool.Query(query, id)
	if err != nil {
		return
	}

	for i := 0; rows.Next(); i++ {
		var g shortGameDTO
		var compressedMoves []int32

		err = rows.Scan(&g.Id, &g.WhiteId, &g.BlackId, &g.TimeControl, &g.TimeBonus,
			&g.Result, &g.Winner, pq.Array(&compressedMoves), &g.CreatedAt,
			&g.WhiteName, &g.BlackName)
		g.MovesLen = len(compressedMoves)
		if err != nil {
			return
		}
		games = append(games, g)
	}
	return
}

func Insert(id string, g chess.Game) error {
	query := "INSERT INTO game (id, white_id, black_id,\n" +
		"time_control, time_bonus, result, winner, moves)\n" +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8);"
	_, err := db.Pool.Exec(query, id, g.WhiteId, g.BlackId,
		g.TimeControl, g.TimeBonus, g.Result, g.Winner,
		pq.Array(compressMoves(g.Moves)))
	return err
}

func compressMoves(moves []chess.CompletedMove) []int {
	compressed := make([]int, len(moves))
	for i, m := range moves {
		compressed[i] = int(m.Move) | (m.TimeLeft << 16)
	}
	return compressed
}

func decompressMoves(moves []int32, fenStr string, control int) []chess.CompletedMove {
	g := chess.NewGame(fenStr, control, 0)

	for i, comp := range moves {
		m := bitboard.Move(comp & 0xFFFF)
		if i%2 == 0 {
			g.WhiteTime = int(comp >> 16)
		} else {
			g.BlackTime = int(comp >> 16)
		}

		g.ProcessMove(m)
	}
	return g.Moves
}
