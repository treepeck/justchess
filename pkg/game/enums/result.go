package enums

type Result uint8

const (
	Unknown Result = iota
	Continues
	Aborted
	Checkmate
	Timeout
	Stalemate
	InsufficienMaterial
	FiftyMoves
	Repetition
	Agreement
)
