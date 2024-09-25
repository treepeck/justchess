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

// client events
const (
	GET_ROOMS           = "GET_ROOMS"
	CREATE_ROOM         = "CREATE_ROOM"
	JOIN_ROOM           = "JOIN_ROOM"
	GET_AVAILIBLE_MOVES = "GET_AVAILIBLE_MOVES"
	MOVE                = "MOVE"
)

// server events
const (
	UPDATE_CLIENTS_COUNTER = "UPDATE_CLIENTS_COUNTER"
	UPDATE_ROOMS           = "UPDATE_ROOMS"
	CHANGE_ROOM            = "CHANGE_ROOM"
	UPDATE_GAME            = "UPDATE_GAME"
	UPDATE_AVAILIBLE_MOVES = "UPDATE_AVAILIBLE_MOVES"
)
