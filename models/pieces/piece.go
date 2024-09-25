package pieces

import (
	"chess-api/enums"
	"chess-api/models/helpers"
)

type Piece interface {
	Move()
	GetName() enums.Piece
	GetColor() enums.Color
	GetPosition() helpers.Position
	// GetAvailibleMoves(map[string]Piece) []helpers.Position
}
