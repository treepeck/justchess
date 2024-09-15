package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Game struct {
	Id      uuid.UUID `json:"id"`
	Control string    `json:"control"` // "blitz" | "bullet" | "rapid"
	Bonus   uint      `json:"bonus"`   // 0 | 1 | 2 | 10
	Status  string    `json:"status"`  // "canceled" | "waiting" | "white_won" |
	// "black_won" | "draw" | "continues"
	WhiteId  uuid.UUID       `json:"whiteId"`
	BlackId  uuid.UUID       `json:"blackId"`
	PlayedAt time.Time       `json:"playedAt"`
	Moves    json.RawMessage `json:"moves"`
}

type CreateGameDTO struct {
	Id      uuid.UUID `json:"id"`
	Control string    `json:"control"`
	Bonus   uint      `json:"bonus"`
	WhiteId uuid.UUID `json:"whiteId"`
	BlackId uuid.UUID `json:"blackId"`
}

func NewGame(cg CreateGameDTO) *Game {
	return &Game{
		Id:      cg.Id,
		Control: cg.Control,
		Bonus:   cg.Bonus,
		Status:  "waiting",
		WhiteId: cg.WhiteId,
		BlackId: cg.BlackId,
	}
}
