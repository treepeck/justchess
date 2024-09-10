package event

import (
	"chess-api/models"
	"encoding/json"
)

type GameEvent struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
	Sender  models.User     `json:"sender"`
}
