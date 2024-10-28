package pieces

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
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
) map[helpers.Pos]enums.MoveType {
	possibleMoves := make(map[helpers.Pos]enums.MoveType)
	traverse(0, 1, pieces, r, possibleMoves)  // upper horizontal (increase rank).
	traverse(0, -1, pieces, r, possibleMoves) // lower horizontal (decrease rank).
	traverse(1, 0, pieces, r, possibleMoves)  // right horizontal (increase file).
	traverse(-1, 0, pieces, r, possibleMoves) // left horizontal (decrease file).
	return possibleMoves
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
