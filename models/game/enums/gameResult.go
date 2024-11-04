package enums

type GameResult int

const (
	Checkmate GameResult = iota
	Resignation
	Timeout
	Stalemate
	InsufficientMaterial
	FiftyMoves
	Repetition
	Agreement
)
