package game

import "justchess/internal/db"

// Move is a payload of [MessageMove].
type Move struct {
	Fen         string         `json:"f"`
	San         string         `json:"s"`
	Result      db.Result      `json:"r,omitempty"`
	Termination db.Termination `json:"t,omitempty"`
	TimeLeft    int            `json:"tl,omitempty"`
	Move        int            `json:"m"`
}
