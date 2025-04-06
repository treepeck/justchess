package game

import (
	"errors"
	"justchess/pkg/chess"
	"justchess/pkg/chess/bitboard"
	"justchess/pkg/chess/enums"
	"justchess/pkg/chess/fen"
	"justchess/pkg/chess/san"
	"justchess/pkg/db"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// shortGameDTO represents a completed game without detailed game info.
type shortGameDTO struct {
	Id          uuid.UUID    `json:"id"`
	Result      enums.Result `json:"r"`
	Winner      enums.Color  `json:"w"`
	MovesLen    int          `json:"m"`
	WhiteId     uuid.UUID    `json:"wid"`
	BlackId     uuid.UUID    `json:"bid"`
	TimeControl int          `json:"tc"`
	TimeBonus   int          `json:"tb"`
}

func selectById(id string) (g chess.Game, err error) {
	query := "SELECT * FROM game WHERE id = $1;"
	rows, err := db.Pool.Query(query, id)
	if err != nil {
		return
	}

	if !rows.Next() {
		return g, errors.New("game not found")
	}
	var compressedMoves []int32
	err = rows.Scan(&g.Id, &g.WhiteId, &g.BlackId, &g.TimeControl, &g.TimeBonus,
		&g.Result, &g.Winner, &g.InitialFEN, pq.Array(&compressedMoves))
	g.Moves = decompressMoves(compressedMoves, g.InitialFEN)
	return
}

func selectByPlayerId(id string) (games []shortGameDTO, err error) {
	query := "SELECT * FROM game WHERE white_id = $1 OR black_id = $1;"
	rows, err := db.Pool.Query(query, id)
	if err != nil {
		return
	}

	for i := 0; rows.Next(); i++ {
		var g shortGameDTO
		var initFEN string
		var compressedMoves []int32
		err = rows.Scan(&g.Id, &g.WhiteId, &g.BlackId, &g.TimeControl, &g.TimeBonus,
			&g.Result, &g.Winner, &initFEN, pq.Array(&compressedMoves))
		g.MovesLen = len(compressedMoves)
		if err != nil {
			return
		}
		games = append(games, g)
	}
	return
}

func Insert(g chess.Game) error {
	query := "INSERT INTO game (id, white_id, black_id, initial_fen,\n" +
		"time_control, time_bonus, result, winner, moves)\n" +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);"
	_, err := db.Pool.Exec(query, g.Id, g.WhiteId, g.BlackId,
		g.InitialFEN, g.TimeControl, g.TimeBonus, g.Result, g.Winner,
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

func decompressMoves(moves []int32, FEN string) []chess.CompletedMove {
	decompressed := make([]chess.CompletedMove, len(moves))
	for i, m := range moves {
		bb := fen.FEN2Bitboard(FEN)
		bb.GenLegalMoves()
		move := bitboard.Move(m & 0xFFFF)
		bb.MakeMove(move)

		FEN = fen.Bitboard2FEN(bb)
		decompressed[i] = chess.CompletedMove{
			Move: move,
			SAN: san.Move2SAN(move, bb.Pieces, bb.LegalMoves,
				bitboard.GetPieceOnSquare(1<<move.To(), bb.Pieces)),
			FEN:      fen.Bitboard2FEN(bb),
			TimeLeft: int(m >> 16),
		}
	}
	return decompressed
}
