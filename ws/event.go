package ws

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
)

type event struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
	Sender  *client         `json:"sender"`
	Target  uuid.UUID       `json:"target"`
}

func (e *event) marshal() []byte {
	json, err := json.Marshal(e)
	if err != nil {
		log.Println("event marshal: ", err)
		return nil
	}
	return json
}

// client events
const (
	CREATE_ROOM = "CREATE_ROOM"
	JOIN_ROOM   = "JOIN_ROOM"
	LEAVE_ROOM  = "LEAVE_ROOM"
)

// server events
const (
	UPDATE_CLIENTS_COUNTER = "UPDATE_CLIENTS_COUNTER"
	UPDATE_ROOMS           = "UPDATE_ROOMS"
)
