package helpers

import (
	"time"

	"github.com/google/uuid"
)

type Player struct {
	Id          uuid.UUID     `json:"id"`
	Time        time.Duration `json:"time"`
	IsConnected bool          `json:"isConnected"`
	Ticker      *time.Ticker  `json:"-"`
	// Extra time is given for reconnection, making the first move.
	// Extra time is equal to 20 seconds by default.
	ExtraTime time.Duration `json:"-"`
}

func NewPlayer(id uuid.UUID, t time.Duration) *Player {
	p := &Player{
		Id:          id,
		Time:        t,
		IsConnected: true,
		Ticker:      time.NewTicker(time.Second),
		ExtraTime:   20 * time.Second,
	}
	p.Ticker.Stop()
	return p
}

func (p *Player) DecrementTime() {
	p.Time -= time.Second
}

func (p *Player) DecrementExtraTime() {
	p.ExtraTime -= time.Second
}
