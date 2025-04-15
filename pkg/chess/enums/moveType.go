package enums

type MoveType int

// MoveType represents all types of moves.
const (
	Quiet MoveType = iota
	DoublePawnPush
	KingCastle  // O-O
	QueenCastle // O-O-O
	Capture
	EPCapture // En passant.
	KnightPromo
	BishopPromo
	RookPromo
	QueenPromo
	KnightPromoCapture
	BishopPromoCapture
	RookPromoCapture
	QueenPromoCapture
)
