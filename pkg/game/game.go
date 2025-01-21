package game

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"justchess/pkg/game/fen"
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
	MoveStack []helpers.Move
	// Piece placement on a board.
	Board *bitboard.Bitboard
	// Game state. Continues by default.
	Status enums.Status
	// False by default.
	IsWhiteKingChecked bool
	IsBlackKingChecked bool
}

func NewGame(bonus uint, timerDur time.Duration) *Game {
	return &Game{
		Bonus:     bonus,
		TimerDur:  timerDur,
		MoveStack: make([]helpers.Move, 0),
		Board:     fen.FEN2Bitboard("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"),
	}
}
