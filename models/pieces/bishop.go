package pieces

import (
	"chess-api/enums"
	"chess-api/models/helpers"
	"log/slog"
)

type Bishop struct {
	Color      enums.Color      `json:"color"`
	IsCaptured bool             `json:"isCaptured"`
	Pos        helpers.Position `json:"pos"`
	Name       enums.Piece      `json:"name"`
}

func NewBishop(color enums.Color, pos helpers.Position) *Bishop {
	return &Bishop{
		Color:      color,
		IsCaptured: false,
		Pos:        pos,
		Name:       enums.Bishop,
	}
}

func (b *Bishop) Move() {
	slog.Debug("Bishop Move")
}

func (b *Bishop) GetName() enums.Piece {
	return enums.Bishop
}

func (b *Bishop) GetColor() enums.Color {
	return b.Color
}

func (b *Bishop) GetPosition() helpers.Position {
	return b.Pos
}
