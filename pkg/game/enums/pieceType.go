package enums

// PieceType describes each possible type of the chess piece.
type PieceType int

const (
	// Any white piece.
	WhiteP PieceType = iota
	// Any black piece.
	BlackP
	// So on.
	WhitePawn
	BlackPawn
	WhiteKnight
	BlackKnight
	WhiteBishop
	BlackBishop
	WhiteRook
	BlackRook
	WhiteQueen
	BlackQueen
	WhiteKing
	BlackKing
)
