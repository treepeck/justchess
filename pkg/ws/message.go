package ws

import (
	"encoding/json"
	"justchess/pkg/game/enums"
)

type MessageType = byte

const (
	CREATE_ROOM MessageType = iota
	MAKE_MOVE
	CHAT
	CLIENTS_COUNTER
	ADD_ROOM
	REMOVE_ROOM
	ROOM_STATUS
	LAST_MOVE
	GAME_RESULT
)

// Message contains Data based on the Type.
type Message struct {
	Type MessageType     `json:"t"`
	Data json.RawMessage `json:"d"`
}

// Client DTOs:
type CreateRoomData struct {
	TimeControl int  `json:"c"`
	TimeBonus   int  `json:"b"`
	IsVSEngine  bool `json:"e"`
}

// Server DTOs:
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

type RoomStatusData struct {
	Status     RoomStatus `json:"s"`
	WhiteId    string     `json:"w"`
	BlackId    string     `json:"b"`
	WhiteTime  int        `json:"wt"`
	BlackTime  int        `json:"bt"`
	IsVSEngine bool       `json:"e"`
	Clients    int        `json:"c"`
}

type LastMoveData struct {
	SAN        string     `json:"s"`
	FEN        string     `json:"f"`
	TimeLeft   int        `json:"t"`
	LegalMoves []MoveData `json:"l"`
}

type GameResultData struct {
	Result enums.Result `json:"r"`
}

// Used by both client and server:
type ChatData struct {
	Message string  `json:"m"`
	client  *client `json:"-"`
}

type MoveData struct {
	To     int            `json:"d"`
	From   int            `json:"s"`
	Type   enums.MoveType `json:"t"`
	client *client        `json:"-"`
}
