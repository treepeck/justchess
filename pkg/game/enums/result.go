package enums

type Result byte

const (
	Unknown Result = iota
	Checkmate
	Timeout
	Stalemate
	InsufficientMaterial
	FiftyMoves
	Repetition
	Agreement
)
