package enums

// MoveType represents all types of moves.
type MoveType uint8

const (
	Quiet MoveType = iota
	DoublePawnPush
	KingCastle  // 0-0
	QueenCastle // 0-0-0
	Capture
	EpCapture // En passant.
	Promotion
)
