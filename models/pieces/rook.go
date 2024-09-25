package pieces

import (
	"chess-api/enums"
	"chess-api/models/helpers"
	"log/slog"
)

type Rook struct {
	Color      enums.Color      `json:"color"`
	IsCaptured bool             `json:"isCaptured"`
	HasMoved   bool             `json:"hasMoved"`
	Pos        helpers.Position `json:"pos"`
	Name       enums.Piece      `json:"name"`
}

func NewRook(color enums.Color, pos helpers.Position) *Rook {
	return &Rook{
		Color:      color,
		IsCaptured: false,
		HasMoved:   false,
		Pos:        pos,
		Name:       enums.Rook,
	}
}

func (r *Rook) Move() {
	slog.Debug("Rook Move")
}

func (r *Rook) GetName() enums.Piece {
	return enums.Rook
}

func (r *Rook) GetColor() enums.Color {
	return r.Color
}

func (r *Rook) GetPosition() helpers.Position {
	return r.Pos
}
