package ws

import (
	"encoding/json"
	"log/slog"
)

type Event struct {
	Action  string          `json:"a"`
	Payload json.RawMessage `json:"p"`
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
	CREATE_ROOM = "cr" // Creates a new room.
	JOIN_ROOM   = "jr" // Join a room.
	LEAVE_ROOM  = "lr" // Leave a room.
	GET_ROOMS   = "gr" // Gets all availible rooms one by one.
	GET_GAME    = "gg" // Get up-to-date game info.
	MOVE        = "m"  // Take a move.
)

// server events
const (
	CLIENTS_COUNTER = "cc" // Updates clients counter.
	ADD_ROOM        = "ar" // Add availible room.
	REMOVE_ROOM     = "rr" // Remove room.
	REDIRECT        = "r"  // Redirect client to a room.
	GAME            = "g"
	UPDATE_BOARD    = "ub" // Redraw board on client.
	MOVES           = "mh" // Moves history.
	STATUS          = "s"  // Up-to-date game status.
	VALID_MOVES     = "vm" // Update valid moves on client.
)

// server errors
const (
	UNPROCESSABLE_ENTITY = "ue"
	CREATE_ROOM_ERR      = "cre"
	JOIN_ROOM_ERR        = "jre"
)
