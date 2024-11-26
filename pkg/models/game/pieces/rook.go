package pieces

import (
	"justchess/pkg/models/game/enums"
	"justchess/pkg/models/game/helpers"
)

type Rook struct {
	Color        enums.Color     `json:"color"`
	MovesCounter uint            `json:"-"`
	Pos          helpers.Pos     `json:"-"`
	Type         enums.PieceType `json:"type"`
}

func NewRook(color enums.Color, pos helpers.Pos) *Rook {
	return &Rook{
		Color:        color,
		MovesCounter: 0,
		Pos:          pos,
		Type:         enums.Rook,
	}
}

func (r *Rook) GetPossibleMoves(pieces map[helpers.Pos]Piece,
) []helpers.PossibleMove {
	pm := make([]helpers.PossibleMove, 0)
	pm = append(pm, traverse(0, 1, pieces, r)...)  // Upper horizontal.
	pm = append(pm, traverse(0, -1, pieces, r)...) // Lower horizontal.
	pm = append(pm, traverse(1, 0, pieces, r)...)  // Right horizontal.
	pm = append(pm, traverse(-1, 0, pieces, r)...) // Left horizontal.
	return pm
}

func (r *Rook) Move(to helpers.Pos) {
	r.Pos = to
	r.MovesCounter++
}

func (r *Rook) GetMovesCounter() uint {
	return r.MovesCounter
}

func (r *Rook) SetMovesCounter(mc uint) {
	r.MovesCounter = mc
}

func (r *Rook) GetType() enums.PieceType {
	return enums.Rook
}

func (r *Rook) GetColor() enums.Color {
	return r.Color
}

func (r *Rook) GetPosition() helpers.Pos {
	return r.Pos
}

func (r *Rook) GetFEN() string {
	if r.Color == enums.White {
		return "R"
	}
	return "r"
}
