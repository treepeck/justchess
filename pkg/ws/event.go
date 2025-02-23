package ws

import "github.com/google/uuid"

type eventType int

const (
	// Client event types.
	REGISTER eventType = iota
	UNREGISTER
	// Game event types.
	GET_AVAILIBLE
	CREATE
	JOIN
	GET
	LEAVE
)

type clientEvent struct {
	eType  eventType
	sender *client
	id     uuid.UUID
}

type gameEvent struct {
	eType   eventType
	payload []byte
}
