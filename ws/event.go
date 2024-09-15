package ws

import (
	"encoding/json"
	"log/slog"
)

type Event struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

func (e *Event) Marshal() []byte {
	json, err := json.Marshal(e)
	if err != nil {
		slog.Warn("game event marshal error", "err", err)
		return nil
	}
	return json
}

type StartGameDTO struct {
	WhiteId string `json:"whiteId"`
	BlackId string `json:"blackId"`
	Control string `json:"control"`
}

// client events
const (
	GET_ROOMS   = "GET_ROOMS"
	CREATE_ROOM = "CREATE_ROOM"
	JOIN_ROOM   = "JOIN_ROOM"
	GAME_ID     = "GAME_ID"
)

// server events
const (
	UPDATE_CLIENTS_COUNTER = "UPDATE_CLIENTS_COUNTER"
	UPDATE_ROOMS           = "UPDATE_ROOMS"
	CHANGE_ROOM            = "CHANGE_ROOM"
	WAITING_OPPONENT       = "WAITING_OPPONENT"
	START_GAME             = "START_GAME"
	OPPONENT_LEFT          = "OPPONENT_LEFT"
	UPDATE_GAME            = "UPDATE_GAME"
)
