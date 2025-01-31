package enums

type Result int

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
