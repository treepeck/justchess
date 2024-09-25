package pieces

import (
	"chess-api/enums"
	"chess-api/models/helpers"
	"log/slog"
)

type King struct {
	Color      enums.Color      `json:"color"`
	IsCaptured bool             `json:"isCaptured"`
	HasMoved   bool             `json:"hasMoved"`
	Pos        helpers.Position `json:"pos"`
	Name       enums.Piece      `json:"name"`
}

func NewKing(color enums.Color, pos helpers.Position) *King {
	return &King{
		Color:      color,
		IsCaptured: false,
		HasMoved:   false,
		Pos:        pos,
		Name:       enums.King,
	}
}

func (k *King) Move() {
	slog.Debug("King Move")
}

func (k *King) GetName() enums.Piece {
	return enums.King
}

func (k *King) GetColor() enums.Color {
	return k.Color
}

func (k *King) GetPosition() helpers.Position {
	return k.Pos
}
