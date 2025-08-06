package ws

import (
	"encoding/json"
	"log"
)

type eventAction int

const (
	actionCounter = iota
	actionCreate  // Create new room.
	actionRemove  // Remove room.
	actionRoomInfo
	actionMakeMove
	actionLastMove
)

type event struct {
	PubId   string          `json:"-"` // Publisher/Sender id.
	TopicId string          `json:"-"` // Topic id.
	Action  eventAction     `json:"a"` // Event type.
	Payload json.RawMessage `json:"p"` // Event payload.
}

// encode is a helper function to encode an event payload.
func encode(payload any) []byte {
	p, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ERROR: cannot encode payload %v", err)
	}
	return p
}

type roomInfo struct {
	Counter int    `json:"c"`
	WhiteId string `json:"w"`
	BlackId string `json:"b"`
}
