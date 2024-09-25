package pieces

import (
	"chess-api/enums"
	"chess-api/models/helpers"
	"log/slog"
)

type Pawn struct {
	Color      enums.Color      `json:"color"`
	HasMoved   bool             `json:"hasMoved"`
	Pos        helpers.Position `json:"pos"`
	Name       enums.Piece      `json:"name"`
	IsCaptured bool             `json:"isCaptured"`
}

func NewPawn(color enums.Color, pos helpers.Position) *Pawn {
	return &Pawn{
		Color:      color,
		HasMoved:   false,
		Pos:        pos,
		Name:       enums.Pawn,
		IsCaptured: false,
	}
}

func (p *Pawn) Move() {
	slog.Debug("Pawn Move")
}

func (p *Pawn) GetName() enums.Piece {
	return enums.Pawn
}

func (p *Pawn) GetColor() enums.Color {
	return p.Color
}

func (p *Pawn) GetPosition() helpers.Position {
	return p.Pos
}
