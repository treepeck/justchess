package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
)

type Piece interface {
	Move(pieces map[helpers.Pos]Piece, to helpers.Pos) bool
	GetName() enums.Piece
	GetColor() enums.Color
	GetPosition() helpers.Pos
	GetAvailibleMoves(pieces map[helpers.Pos]Piece) []helpers.Pos
}
