package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
	"log/slog"
)

type Queen struct {
	Color      enums.Color `json:"color"`
	IsCaptured bool        `json:"isCaptured"`
	Pos        helpers.Pos `json:"pos"`
	Name       enums.Piece `json:"name"`
}

func NewQueen(color enums.Color, pos helpers.Pos) *Queen {
	return &Queen{
		Color:      color,
		IsCaptured: false,
		Pos:        pos,
		Name:       enums.Queen,
	}
}

func (q *Queen) Move(pieces map[helpers.Pos]Piece, to helpers.Pos) bool {
	slog.Debug("Queen Move")
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

func (q *Queen) GetAvailibleMoves(map[helpers.Pos]Piece) []helpers.Pos {
	return make([]helpers.Pos, 0)
}
