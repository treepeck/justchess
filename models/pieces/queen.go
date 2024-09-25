package pieces

import (
	"chess-api/enums"
	"chess-api/models/helpers"
	"log/slog"
)

type Queen struct {
	Color      enums.Color      `json:"color"`
	IsCaptured bool             `json:"isCaptured"`
	Pos        helpers.Position `json:"pos"`
	Name       enums.Piece      `json:"name"`
}

func NewQueen(color enums.Color, pos helpers.Position) *Queen {
	return &Queen{
		Color:      color,
		IsCaptured: false,
		Pos:        pos,
		Name:       enums.Queen,
	}
}

func (q *Queen) Move() {
	slog.Debug("Queen Move")
}

func (q *Queen) GetName() enums.Piece {
	return enums.Queen
}

func (q *Queen) GetColor() enums.Color {
	return q.Color
}

func (q *Queen) GetPosition() helpers.Position {
	return q.Pos
}
