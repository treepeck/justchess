package core

import "github.com/treepeck/chego"

type moveDTO struct {
	playerId string
	move     chego.Move
}

/*
roomState represents a domain of possible room states.
*/
type roomState int

const (
	// stateEmpty means that no clients are connected.
	stateEmpty roomState = iota
	// stateWhite means that only white player is connected.
	stateWhite
	// stateBlack means that only black player is connected.
	stateBlack
	// stateBoth means that both players are connected.
	stateBoth
)
