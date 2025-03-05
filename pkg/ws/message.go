package ws

import (
	"encoding/json"
)

type MessageType = byte

const (
	// Sent by clients.
	CREATE_ROOM MessageType = iota
	MAKE_MOVE

	// Sent by server.
	CLIENTS_COUNTER
	ADD_ROOM
	REMOVE_ROOM
	CHAT_MESSAGE
	ROOM_INFO
	GAME
	LAST_MOVE
	RESULT
)

// Message contains Data based on the Type.
type Message struct {
	Type MessageType     `json:"t"`
	Data json.RawMessage `json:"d"`
}

type ClientsCounterData struct {
	Counter int `json:"c"`
}

type AddRoomData struct {
	CreatorId   string `json:"id"`
	TimeControl int    `json:"c"`
	TimeBonus   int    `json:"b"`
}

type RemoveRoomData struct {
	RoomId string `json:"id"`
}

type CreateRoomData struct {
	TimeControl int `json:"c"`
	TimeBonus   int `json:"b"`
}

type MakeMoveData struct {
	To   int `json:"to"`
	From int `json:"from"`
	Type int `json:"type"`
}
