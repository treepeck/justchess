package enums

// Status represents a game states.
type Status int

const (
	// The basic state of the game before it is completed.
	Continues Status = iota
	// White player delivers checkmate.
	WhiteWon
	// Black player delivers checkmate.
	BlackWon
	// White player resigns.
	WhiteResign
	// Black player resigns.
	BlackResign
	// White player runs out of time.
	WhiteTimeout
	// Black player runs out of time.
	BlackTimeout
	// Draw by stalemate.
	Stalemate
	// Both players does not have enouth pieces to deliver a checkmate. Draw.
	InsufficientMaterial
	// No capture has been made and no pawn has been moved in the last 50 moves. Draw.
	FiftyMoves
	// The same position reached 3 times. Draw.
	Repetition
	// Draw by player`s agreement.
	Agreement
)
