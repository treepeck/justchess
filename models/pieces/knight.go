package pieces

import (
	"chess-api/enums"
	"chess-api/models/helpers"
	"log/slog"
)

type Knight struct {
	Color      enums.Color      `json:"color"`
	IsCaptured bool             `json:"isCaptured"`
	Pos        helpers.Position `json:"pos"`
	Name       enums.Piece      `json:"name"`
}

func NewKnight(color enums.Color, pos helpers.Position) *Knight {
	return &Knight{
		Color:      color,
		IsCaptured: false,
		Pos:        pos,
		Name:       enums.Knight,
	}
}

func (k *Knight) Move() {
	slog.Debug("Knight Move")
}

func (k *Knight) GetName() enums.Piece {
	return enums.Knight
}

func (k *Knight) GetColor() enums.Color {
	return k.Color
}

func (k *Knight) GetPosition() helpers.Position {
	return k.Pos
}
