package enums

// PieceType describes each possible type of the chess piece.
type PieceType int

const (
	// Any piece.
	Piece PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)
