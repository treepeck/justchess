package enums

type PieceType = int

const (
	Piece PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)
