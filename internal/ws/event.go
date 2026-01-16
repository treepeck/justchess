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
	actionMakeMove

	// Server's actions.
	actionClientsCounter
	actionRedirect
)

type event struct {
	Payload json.RawMessage `json:"p"`
	Action  eventAction     `json:"a"`
	// Ignored in json.
	sender *client
}
