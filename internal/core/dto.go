package core

import (
	"github.com/treepeck/chego"
)

type joinMatchmakingReq struct {
	playerId string
	params   roomParams
}

type addRoomRes struct {
	players     [2]string
	roomId      string
	timeControl int
	timeBonus   int
}

type moveReq struct {
	playerId string
	move     chego.Move
}

type roomInfo struct {
	WhiteId string `json:"w"`
	BlackId string `json:"b"`
	// In seconds.
	TimeToLive int `json:"t"`
	Viewers    int `json:"v"`
}

type gameState struct {
	CompletedMoves []completedMove `json:"cm"`
	LegalMoves     []chego.Move    `json:"lm"`
	WhiteTime      int             `json:"w"`
	BlackTime      int             `json:"b"`
}

type roomParams struct {
	TimeControl int `json:"tc"`
	TimeBonus   int `json:"tb"`
}

type completedMove struct {
	San  string     `json:"s"`
	Move chego.Move `json:"m"`
}
