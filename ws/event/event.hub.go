package event

import (
	"encoding/json"
	"log"
)

type HubEvent struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

func (e *HubEvent) Marshal() []byte {
	json, err := json.Marshal(e)
	if err != nil {
		log.Println("event marshal: ", err)
		return nil
	}
	return json
}

// client events
const (
	GET_ROOMS   = "GET_ROOMS"
	CREATE_ROOM = "CREATE_ROOM"
	JOIN_ROOM   = "JOIN_ROOM"
)

// server events
const (
	UPDATE_CLIENTS_COUNTER = "UPDATE_CLIENTS_COUNTER"
	UPDATE_ROOMS           = "UPDATE_ROOMS"
	CHANGE_ROOM            = "CHANGE_ROOM"
)
