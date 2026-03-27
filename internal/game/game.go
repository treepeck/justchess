// Package game implements real time game management.
package game

import "github.com/treepeck/chego"

const (
	// Minimal number of moves required to terminate the game.
	// Otherwise the game is marked as abandoned.
	minMoves = 3

	// Disconnected players have 30 seconds to reconnect.  If the player doesn't
	// reconnect within the specified time period, victory is awarded to the
	// opponent.
	reconnectDeadline = 30
)

// Game methods are not safe for concurrent use.
type Game interface {
	Play(id string, index byte) (MovePayload, bool)
	// Resign handles player resignation.  Resignation will be discarded if one
	// of the following is true:
	//   - There were not enough moves played to end the game;
	//   - The game is already over;
	//   - Sender is not white nor black player.
	Join(id string)
	Leave(id string)
	TimeTick()
	Resign(id string) bool
	GamePayload() GamePayload
	EndPayload() EndPayload
	Abandon()
}

type GamePayload struct {
	Legal  []chego.Move       `json:"lm"`
	Played []chego.PlayedMove `json:"m"`
	// Clock values in seconds if present.
	WhiteTime int `json:"wt,omitempty"`
	BlackTime int `json:"bt,omitempty"`
}

type EndPayload struct {
	Result      chego.Result      `json:"r"`
	Termination chego.Termination `json:"t"`
}

type MovePayload struct {
	chego.PlayedMove
	Legal    []chego.Move `json:"lm"`
	TimeLeft int          `json:"tl,omitempty"`
}
