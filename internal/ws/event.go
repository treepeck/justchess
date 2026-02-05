package ws

import (
	"encoding/json"

	"github.com/treepeck/chego"
)

// Domain of possible event actions.
type eventAction int

const (
	// Ping is sent by the server to maintain a heartbeat and detect idle
	// connections.  The payload contains the network latency in milliseconds.
	actionPing eventAction = iota
	// Pong must be sent by the client immediately after receiving a [actionPing].
	actionPong
	// Chat represents a chat message sent by a client or broadcast by a [room].
	actionChat
	// Move represents a chess move performed by a client or broadcast by a [room].
	actionMove
	// Game represents the current game state sent to each client after
	// connecting to a [room], allowing them to synchronize.
	actionGame
	// Conn is broadcast by a [room] to notify clients about a player connection.
	actionConn
	// Disc is broadcast by a [room] to notify clients about a player disconnection.
	actionDisc
	// ClientsCounter is broadcast in a [queue] whenever a player joins or leaves.
	actionClientsCounter
	// Redirect is sent to players after a match is found, redirecting them
	// to the game [room].
	actionRedirect
	// Error contains an error message payload.
	actionError
)

// event represents an arbitrary event.
type event struct {
	Payload json.RawMessage `json:"p"`
	Action  eventAction     `json:"a"`
	// Ignored in json.
	sender *client
}

// encodes the given payload and event.
func newEncodedEvent(a eventAction, payload any) ([]byte, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(event{
		Action:  a,
		Payload: p,
	})
}

type completedMove struct {
	San  string     `json:"s"`
	Move chego.Move `json:"m"`
	// Remaining time on the player's clock.
	TimeLeft int `json:"t"`
}

// gamePayload is a payload for the event with [Game] action.
type gamePayload struct {
	LegalMoves []chego.Move    `json:"lm"`
	Moves      []completedMove `json:"m"`
	WhiteTime  int             `json:"wt"`
	BlackTime  int             `json:"bt"`
}

type movePayload struct {
	LegalMoves []chego.Move  `json:"lm"`
	Move       completedMove `json:"m"`
}
