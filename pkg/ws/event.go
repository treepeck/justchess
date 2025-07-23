package ws

import "encoding/json"

type eventAction int

const (
	actionCounter = iota
	actionCreate  // Create new room.
	actionRemove  // Remove room.
)

type event struct {
	PubId   string          `json:"-"` // Publisher/Sender id.
	TopicId string          `json:"-"` // Topic id.
	Action  eventAction     `json:"a"` // Event type.
	Payload json.RawMessage `json:"p"` // Event payload.
}
