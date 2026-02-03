package ws

import (
	"encoding/json"

	"github.com/treepeck/chego"
)

// Domain of possible event actions.
type eventAction int

const (
	actionPing eventAction = iota
	actionPong
	actionChat
	actionMove
	actionGame
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
	res chan *room
}

type gamePayload struct {
	LegalMoves       []chego.Move    `json:"lm"`
	Moves            []completedMove `json:"m"`
	WhiteTime        int             `json:"wt"`
	BlackTime        int             `json:"bt"`
	IsWhiteConnected bool            `json:"w"`
	IsBlackConnected bool            `json:"b"`
}

type movePayload struct {
	LegalMoves []chego.Move  `json:"lm"`
	Move       completedMove `json:"m"`
}
