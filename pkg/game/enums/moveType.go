package enums

// MoveType represents all types of moves.
const (
	Quiet int = iota
	DoublePawnPush
	KingCastle  // 0-0
	QueenCastle // 0-0-0
	Capture
	EpCapture // En passant.
	Promotion
)
