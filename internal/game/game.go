package game

import (
	"crypto/rand"

	"github.com/BelikovArtem/chego"
)

type GameRoom struct {
	Id      string
	WhiteId string
	BlackId string
	game    *chego.Game
}

func NewGameRoom(timeControl, timeBonus int) *GameRoom {
	gr := &GameRoom{
		Id:   rand.Text(),
		game: chego.NewGame(),
	}

	return gr
}
