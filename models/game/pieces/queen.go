package pieces

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
)

type Queen struct {
	Color        enums.Color     `json:"color"`
	Pos          helpers.Pos     `json:"-"`
	Type         enums.PieceType `json:"type"`
	MovesCounter uint            `json:"-"`
}

func NewQueen(color enums.Color, pos helpers.Pos) *Queen {
	return &Queen{
		Color:        color,
		Pos:          pos,
		Type:         enums.Queen,
		MovesCounter: 0,
	}
}

func (q *Queen) GetPossibleMoves(pieces map[helpers.Pos]Piece,
) []helpers.PossibleMove {
	pm := make([]helpers.PossibleMove, 0)
	// queen`s move pattern is simply an addition of the rook and the bishop moves.
	rook := NewRook(q.Color, q.Pos)
	bishop := NewBishop(q.Color, q.Pos)
	pm = append(pm, rook.GetPossibleMoves(pieces)...)
	pm = append(pm, bishop.GetPossibleMoves(pieces)...)
	return pm
}

func (q *Queen) Move(to helpers.Pos) {
	q.Pos = to
	q.MovesCounter++
}

func (q *Queen) GetMovesCounter() uint {
	return q.MovesCounter
}

func (q *Queen) SetMovesCounter(mc uint) {
	q.MovesCounter = mc
}

func (q *Queen) GetType() enums.PieceType {
	return enums.Queen
}

func (q *Queen) GetColor() enums.Color {
	return q.Color
}

func (q *Queen) GetPosition() helpers.Pos {
	return q.Pos
}

func (q *Queen) GetFEN() string {
	if q.Color == enums.White {
		return "Q"
	}
	return "q"
}
