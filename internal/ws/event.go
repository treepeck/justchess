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
	actionJoinMatchmaking
	actionLeaveMatchmaking
	actionMakeMove
	actionJoin
	actionLeave

	// Room's actions.
	actionRedirect
)

// Recieved from or forwared to the client struct.
type clientEvent struct {
	Payload json.RawMessage `json:"p"`
	Action  eventAction     `json:"a"`
	sender  *client         `json:"-"`
}

type roomEvent struct {
	recipients []string
	payload    json.RawMessage
	action     eventAction
}
