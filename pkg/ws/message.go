package ws

import "github.com/BelikovArtem/chego/enum"

// messageAction is a domain of all possible message types.
type messageAction int

const (
	actionRoomInfo messageAction = iota
	actionMakeMove
	actionLastMove
	actionGameOver
)

// message represents a single message exchanged between the client and server.
// All data is encoded into JSON.
// Payload is a nested JSON whose actual type depends on the message action
type message struct {
	Payload string        `json:"p"`
	Action  messageAction `json:"a"`
}

// roomInfoPayload is payload type of each message with the RoomInfo action.
// It gives the client information about the room.
type roomInfoPayload struct {
	State            roomState `json:"rs"`
	ClientsCounter   int       `json:"cc"`
	IsWhiteConnected bool      `json:"w"`
	IsBlackConnected bool      `json:"b"`
}

// makeMovePayload is payload type of each message with the MakeMove action.
type makeMovePayload struct {
	To             int        `json:"t"`
	From           int        `json:"f"`
	PromotionPiece enum.Piece `json:"p"`
	senderId       string     `json:"-"`
}

type lastMovePayload struct {
	LegalMoves map[string]string `json:"l"`
	// Game state after completing the move.
	FenString string `json:"f"`
}

type gameOverPayload struct {
	Result enum.Result `json:"r"`
}
