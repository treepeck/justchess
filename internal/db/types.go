package db

import "time"

// Player represents a registered player.  Sensitive data, such as password hash
// and email will not be encoded into a JSON.
type Player struct {
	PasswordHash     []byte    `json:"-"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Id               string    `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"-"`
	Rating           float64   `json:"rating"`
	RatingDeviation  float64   `json:"-"`
	RatingVolatility float64   `json:"-"`
}

// Session is an authorization token for a player.  Each protected endpoint
// expects the Auth cookie to contain valid and not expired session.
type Session struct {
	CreatedAt time.Time
	ExpiresAt time.Time
	Id        string
	PlayerId  string
}

// Game represents the state of a single completed chess game.
type Game struct {
	CreatedAt   time.Time
	Id          string      `json:"id"`
	WhiteId     string      `json:"whiteId"`
	BlackId     string      `json:"blackId"`
	Control     int         `json:"control"`
	Bonus       int         `json:"bonus"`
	Result      Result      `json:"result"`
	Termination Termination `json:"termination"`
}

// Result represents the possible outcomes of a chess game.
type Result int

const (
	Unknown Result = iota
	WhiteWon
	BlackWon
	Draw
)

// Termination represents the reason for the conclusion of the game.  While the
// [Result] types gives the result of the game, it does not provide any extra
// information and so the Termination type is defined for this purpose.
type Termination int

const (
	Unterminated Termination = iota
	Abandoned
	Adjudication
	Normal
	RulesInfraction
	TimeForfeit
)
