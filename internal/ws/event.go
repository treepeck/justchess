package ws

import (
	"encoding/json"
)

// Domain of possible event actions.
type eventAction int

const (
	// Client's actions.
	actionPing eventAction = iota
	actionPong
	actionChat
	actionMove

	// Server's actions.
	actionClientsCounter
	actionRedirect
	actionError
)

type event struct {
	Payload json.RawMessage `json:"p"`
	Action  eventAction     `json:"a"`
	// Ignored in json.
	sender *client
}

type createRoomEvent struct {
	id      string
	whiteId string
	blackId string
	control int
	bonus   int
	res     chan error
}

type findRoomEvent struct {
	id  string
	res chan room
}
