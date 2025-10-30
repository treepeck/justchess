package core

import "github.com/treepeck/chego"

type moveDTO struct {
	playerId string
	move     chego.Move
}

type matchmakingDTO struct {
	TimeControl int `json:"tc"`
	TimeBonus   int `json:"tb"`
}
