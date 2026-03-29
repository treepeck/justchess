package event

import (
	"encoding/json"
	"log"
)

type Kind int

const (
	// Ping is sent by the server to maintain a heartbeat and detect idle
	// connections.  The payload contains the network latency in milliseconds.
	Ping Kind = iota
	// Pong must be sent by the client immediately after receiving a [Ping].
	Pong
	Chat
	Resign
	OfferDraw
	AcceptDraw
	DeclineDraw
	Move
	Game
	End
	Conn
	Disc
	ClientsCounter
	Redirect
	Error
)

type Event struct {
	Payload  json.RawMessage `json:"p"`
	Kind     Kind            `json:"k"`
	SenderId string          `json:"-"`
}

// JSON returns encoded event.  Errors are ignored.
func JSON(k Kind, p any) json.RawMessage {
	rawPayload, err := json.Marshal(p)
	if err != nil {
		log.Print(err)
	}
	rawEvent, err := json.Marshal(Event{Kind: k, Payload: rawPayload})
	if err != nil {
		log.Print(err)
	}
	return rawEvent

}
