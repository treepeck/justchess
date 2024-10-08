package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
)

type Queen struct {
	Color enums.Color `json:"color"`
	Pos   helpers.Pos `json:"pos"`
	Name  enums.Piece `json:"name"`
}

func NewQueen(color enums.Color, pos helpers.Pos) *Queen {
	return &Queen{
		Color: color,
		Pos:   pos,
		Name:  enums.Queen,
	}
}

func (q *Queen) Move(pieces map[helpers.Pos]Piece, move *helpers.Move) bool {
	possibleMoves := q.GetPossibleMoves(pieces)

	pm := possibleMoves[move.To]
	if pm != 0 && pm != enums.Defend {
		if pieces[move.To] != nil {
			move.IsCapture = true
		}

		delete(pieces, move.From)
		pieces[move.To] = q
		q.Pos = move.To
		return true
	}

	return false
}

func (q *Queen) GetName() enums.Piece {
	return enums.Queen
}

func (q *Queen) GetColor() enums.Color {
	return q.Color
}

func (q *Queen) GetPosition() helpers.Pos {
	return q.Pos
}

func (q *Queen) GetPossibleMoves(pieces map[helpers.Pos]Piece,
) map[helpers.Pos]enums.MoveType {
	possibleMoves := make(map[helpers.Pos]enums.MoveType)

	// queen moves is just a concatenation of the rook and the bishop moves
	rook := NewRook(q.Color, q.Pos)
	bishop := NewBishop(q.Color, q.Pos)

	for pos, mt := range rook.GetPossibleMoves(pieces) {
		possibleMoves[pos] = mt
	}
	for pos, mt := range bishop.GetPossibleMoves(pieces) {
		possibleMoves[pos] = mt
	}

	return possibleMoves
}
