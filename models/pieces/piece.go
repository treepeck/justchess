package pieces

import (
	"chess-api/models/enums"
	"chess-api/models/helpers"
)

type Piece interface {
	Move(pieces map[helpers.Pos]Piece, move *helpers.Move) bool
	GetName() enums.Piece
	GetColor() enums.Color
	GetPosition() helpers.Pos
	GetPossibleMoves(pieces map[helpers.Pos]Piece) map[helpers.Pos]enums.MoveType
}
