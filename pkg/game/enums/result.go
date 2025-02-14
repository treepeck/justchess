package enums

type Result uint8

const (
	Continues Result = iota
	Checkmate
	Timeout
	Stalemate
	InsufficienMaterial
	FiftyMoves
	Repetition
	Agreement
	Unknown
)
