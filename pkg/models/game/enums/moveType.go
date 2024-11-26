package enums

type MoveType int

const (
	Basic MoveType = iota + 1 // move to an empty square
	PawnForward
	Defend // forbits the enemy king to move on an attacked square
	LongCastling
	ShortCastling
	EnPassant
	Promotion
)
