package enums

type Result byte

const (
	Unknown Result = iota
	Checkmate
	Timeout
	Stalemate
	InsufficienMaterial
	FiftyMoves
	Repetition
	Agreement
)
