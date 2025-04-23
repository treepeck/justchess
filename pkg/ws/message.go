package ws

import (
	"encoding/json"
	"justchess/pkg/chess/bitboard"
	"justchess/pkg/chess/enums"

	"github.com/google/uuid"
)

type MessageType = byte

const (
	CREATE_ROOM MessageType = iota
	MAKE_MOVE
	CHAT
	RESIGN
	DRAW_OFFER
	DECLINE_DRAW
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
type CreateRoomDTO struct {
	TimeControl int  `json:"c"`
	TimeBonus   int  `json:"b"`
	IsVSEngine  bool `json:"e"`
}

type MoveDTO struct {
	Destination int            `json:"d"`
	Source      int            `json:"s"`
	Type        enums.MoveType `json:"t"`
}

// Server DTOs:
type ClientsCounterDTO struct {
	Counter int `json:"c"`
}

type AddRoomDTO struct {
	Id          uuid.UUID `json:"id"`
	Creator     string    `json:"cr"`
	TimeControl int       `json:"c"`
	TimeBonus   int       `json:"b"`
}

type RemoveRoomDTO struct {
	RoomId string `json:"id"`
}

type RoomStatusDTO struct {
	Status     roomStatus `json:"s"`
	White      uuid.UUID  `json:"w"`
	Black      uuid.UUID  `json:"b"`
	Control    int        `json:"tc"`
	WhiteTime  int        `json:"wt"`
	BlackTime  int        `json:"bt"`
	IsVSEngine bool       `json:"e"`
	Clients    int        `json:"c"`
}

type LastMoveDTO struct {
	SAN        string          `json:"s"`
	FEN        string          `json:"f"`
	TimeLeft   int             `json:"t"`
	LegalMoves []bitboard.Move `json:"l"`
}

type GameResultDTO struct {
	Result enums.Result `json:"r"`
	Winner enums.Color  `json:"w"`
}

// Used by both client and server:
type ChatDTO struct {
	Message string `json:"m"`
}
