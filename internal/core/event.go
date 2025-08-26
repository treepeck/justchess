package core

import "github.com/BelikovArtem/gatekeeper/pkg/event"

const (
	// Client events.
	MAKE_MOVE event.EventAction = 6
	CHAT      event.EventAction = 7
)

type createRoomPayload struct {
	TimeControl int `json:"tc"`
	TimeBonus   int `json:"tb"`
}

type addRoomPayload struct {
	Id          string `json:"id"`
	CreatorId   string `json:"cid"`
	TimeControl int    `json:"tc"`
	TimeBonus   int    `json:"tb"`
}
