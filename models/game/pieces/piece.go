package pieces

import (
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
)

// Piece describes the methods, that all concrete pieces implement.
type Piece interface {
	// GetPossibleMoves finds all possible moves for the piece.
	// The validity of the returned moves is not guaranteed.
	// GetPossibleMoves checks only the following conditions:
	// 	 1. Move position is not occupied by the allied piece;
	// 	 2. Move position corresponds to the piece`s movement pattern.
	// That means, each move returned by the GetPossibleMoves must be additionaly
	// checked for:
	//   1. Making this move does not expose the allied king to check;
	//   2. If the allied king is checked, the move is valid only if it blocks the
	//      king from the check.
	// The map is used as a returned type to simplify the testing
	// (there is no need to specify the exact order of the move positions).
	GetPossibleMoves(pieces map[helpers.Pos]Piece) map[helpers.Pos]enums.MoveType
	// Move moves the piece to a new position.
	// Move does not modify the board state [chess-api/models/Game].
	// The validity of a move must be checked before calling.
	Move(to helpers.Pos)
	// GetMovesCounter returns the number of piece moves.
	GetMovesCounter() uint
	// SetMovesCounter sets the number of piece moves.
	SetMovesCounter(mc uint)
	// GetPosition returns the piece position.
	GetPosition() helpers.Pos
	// GetType returns the piece type.
	GetType() enums.PieceType
	// GetColor returns the piece color.
	GetColor() enums.Color
}

// BuildPiece takes concrete piece and returns that piece copy.
func BuildPiece(piece Piece) Piece {
	var p Piece
	switch piece.GetType() {
	case enums.Pawn:
		p = NewPawn(piece.GetColor(), piece.GetPosition())
	case enums.Knight:
		p = NewKnight(piece.GetColor(), piece.GetPosition())
	case enums.Bishop:
		p = NewBishop(piece.GetColor(), piece.GetPosition())
	case enums.Rook:
		p = NewRook(piece.GetColor(), piece.GetPosition())
	case enums.Queen:
		p = NewQueen(piece.GetColor(), piece.GetPosition())
	case enums.King:
		p = NewKing(piece.GetColor(), piece.GetPosition())
	}
	p.SetMovesCounter(piece.GetMovesCounter())
	return p
}
