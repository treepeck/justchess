package pieces

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
)

type Bishop struct {
	Color        enums.Color     `json:"color"`
	Pos          helpers.Pos     `json:"-"`
	Type         enums.PieceType `json:"type"`
	MovesCounter uint            `json:"-"`
}

func NewBishop(color enums.Color, pos helpers.Pos) *Bishop {
	return &Bishop{
		Color:        color,
		Pos:          pos,
		Type:         enums.Bishop,
		MovesCounter: 0,
	}
}

func (b *Bishop) GetPossibleMoves(pieces map[helpers.Pos]Piece,
) map[helpers.Pos]enums.MoveType {
	possibleMoves := make(map[helpers.Pos]enums.MoveType)
	traverse(-1, 1, pieces, b, possibleMoves)  // Upper left diagonal (decrease file, increase rank).
	traverse(-1, -1, pieces, b, possibleMoves) // Lower left diagonal (decrease file, decrease rank).
	traverse(1, 1, pieces, b, possibleMoves)   // Upper right diagonal (increase file, increase rank).
	traverse(1, -1, pieces, b, possibleMoves)  // Lower right diagonal (increase file, decrease rank).
	return possibleMoves
}

func (b *Bishop) Move(to helpers.Pos) {
	b.Pos = to
	b.MovesCounter++
}

func (b *Bishop) GetMovesCounter() uint {
	return b.MovesCounter
}

func (b *Bishop) SetMovesCounter(mc uint) {
	b.MovesCounter = mc
}

func (b *Bishop) GetType() enums.PieceType {
	return enums.Bishop
}

func (b *Bishop) GetColor() enums.Color {
	return b.Color
}

func (b *Bishop) GetPosition() helpers.Pos {
	return b.Pos
}
