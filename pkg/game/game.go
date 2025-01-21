package game

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"justchess/pkg/game/helpers"
	"time"
)

// Game type represents a chess match.
type Game struct {
	// The number of seconds added after each move.
	Bonus uint
	// Initial amount of time on player`s timers.
	TimerDur time.Duration
	// Completed moves in a historical order.
	Moves []helpers.Move
	// Piece placement on a board.
	Board *bitboard.Bitboard
	// Which side is currently moving. White by default.
	ActiveColor enums.Color
	// Game state. Continues by default.
	Status enums.Status
	// False by default.
	IsWhiteKingChecked bool
	IsBlackKingChecked bool
}

func NewGame(bonus uint, timerDur time.Duration) *Game {
	return &Game{
		Bonus:    bonus,
		TimerDur: timerDur,
		Moves:    make([]helpers.Move, 0),
		Board:    bitboard.NewBitboard(),
	}
}
