package pieces

import (
	"justchess/pkg/models/game/enums"
	"justchess/pkg/models/game/helpers"
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
) []helpers.PossibleMove {
	pm := make([]helpers.PossibleMove, 0)
	pm = append(pm, traverse(-1, 1, pieces, b)...)  // Upper left diagonal.
	pm = append(pm, traverse(-1, -1, pieces, b)...) // Lower left diagonal.
	pm = append(pm, traverse(1, 1, pieces, b)...)   // Upper right diagonal.
	pm = append(pm, traverse(1, -1, pieces, b)...)  // Lower right diagonal.
	return pm
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

func (b *Bishop) GetFEN() string {
	if b.Color == enums.White {
		return "B"
	}
	return "b"
}
