package talk

import "encoding/json"

type MessageKind int

const (
	MessagePing MessageKind = iota
	MessagePong
	MessageClientsCounter
	MessageMove
	MessageOfferDraw
	MessageDeclineDraw
	MessageAcceptDraw
	MessageResign
	MessageError
	MessageRedirect
	MessageJoin
	MessageLeave
)

type Message struct {
	Payload  json.RawMessage `json:"p"`
	Kind     MessageKind     `json:"k"`
	PlayerId string          `json:"-"`
}

func JSON(k MessageKind, p any) ([]byte, error) {
	raw, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Message{Kind: k, Payload: raw})
}
