package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
	"log/slog"
)

type King struct {
	Color        enums.Color `json:"color"`
	IsCaptured   bool        `json:"isCaptured"`
	MovesCounter uint        `json:"movesCounter"`
	Pos          helpers.Pos `json:"pos"`
	Name         enums.Piece `json:"name"`
}

func NewKing(color enums.Color, pos helpers.Pos) *King {
	return &King{
		Color:        color,
		IsCaptured:   false,
		MovesCounter: 0,
		Pos:          pos,
		Name:         enums.King,
	}
}

func (k *King) Move(pieces map[helpers.Pos]Piece, to helpers.Pos) bool {
	slog.Debug("King Move")
	return false
}

func (k *King) GetName() enums.Piece {
	return enums.King
}

func (k *King) GetColor() enums.Color {
	return k.Color
}

func (k *King) GetPosition() helpers.Pos {
	return k.Pos
}

func (k *King) GetMovesCounter() uint {
	return k.MovesCounter
}

func (k *King) GetAvailibleMoves(map[helpers.Pos]Piece) []helpers.Pos {
	return make([]helpers.Pos, 0)
}
