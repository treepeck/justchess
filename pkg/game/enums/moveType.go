package enums

type MoveType int

// MoveType represents all types of moves.
const (
	Quiet MoveType = iota
	DoublePawnPush
	KingCastle  // 0-0
	QueenCastle // 0-0-0
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
