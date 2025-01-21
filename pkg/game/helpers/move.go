package helpers

import (
	"justchess/pkg/game/enums"
)

type Move struct {
	To   int
	From int
	// Describes the board state after making the move.
	FEN               string
	Color             enums.Color
	MoveType          enums.MoveType
	PieceType         enums.PieceType
	CapturedPieceType enums.PieceType
}

func NewMove(to, from int, mt enums.MoveType) Move {
	return Move{
		To:       to,
		From:     from,
		MoveType: mt,
	}
}
