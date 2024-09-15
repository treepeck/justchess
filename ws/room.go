package ws

import (
	"chess-api/models"
	"encoding/json"
	"log/slog"
	"math/rand"
	"sync"

	"github.com/google/uuid"
)

type Room struct {
	sync.Mutex
	Id          uuid.UUID   `json:"id"`
	Owner       models.User `json:"owner"`
	WhitePlayer *Client
	BlackPlayer *Client
	Control     string `json:"control"` // "blitz" | "bullet" | "rapid"
	Bonus       uint   `json:"bonus"`   // 0 | 1 | 2 | 10
	IsAvailible bool   `json:"isAvailible"`
}

type CreateRoomDTO struct {
	// Control must be "blitz", "bullet" or "rapid". See schema.sql.
	Control string `json:"control"`
	// Bonus must be 0, 1, 2 or 10. See schema.sql.
	Bonus  uint        `json:"bonus"`
	Rating uint        `json:"rating"`
	Owner  models.User `json:"owner"`
}

func NewRoom(cr CreateRoomDTO, owner *Client) *Room {
	return &Room{
		Id:          uuid.New(),
		Owner:       cr.Owner,
		WhitePlayer: nil,
		BlackPlayer: nil,
		Control:     cr.Control,
		Bonus:       cr.Bonus,
		IsAvailible: true,
	}
}

func (r *Room) AddPlayer(c *Client) {
	r.Lock()
	defer r.Unlock()

	fn := slog.String("func", "AddPlayer")
	if r.IsAvailible {
		if r.WhitePlayer == nil && r.BlackPlayer == nil {
			// first player takes up random side
			if rand.Intn(2) == 1 {
				r.WhitePlayer = c
				slog.Info("white player connected", fn)
			} else {
				r.BlackPlayer = c
				slog.Info("black player connected", fn)
			}

			// send WAITING_OPPONENT event back to the client
			e := Event{
				Action:  WAITING_OPPONENT,
				Payload: nil,
			}
			c.writeEventBuffer <- e
		} else {
			slog.Debug("players", "white", r.WhitePlayer, "black", r.BlackPlayer)
			// second player takes up the availible side
			if r.WhitePlayer == nil {
				r.WhitePlayer = c
				slog.Info("white player connected", fn)
			} else {
				r.BlackPlayer = c
				slog.Info("black player connected", fn)
			}
			// make the room unavailible to other clients
			r.IsAvailible = false

			// broadcast START_GAME event
			r.Broadcast(START_GAME)
		}
	}
}

// func (r *Room) AddPlayer(c *Client) {
// 	if r.IsAvailible {
// 		if r.Players[0] == nil {
// 			r.Players[0] = c
// 		} else if r.Players[1] == nil {
// 			r.Players[1] = c
// 		}
// 		// close the room
// 		r.IsAvailible = false
// 	}
// }

// func (r *Room) RemovePlayer(c *Client) {
// 	if r.Players[0] == c {

// 	} else if r.BlackPlayer == c {

// 	}
// }

func (r *Room) Broadcast(action string) {
	fn := slog.String("func", "room.broadcast")

	var e Event
	switch action {

	case START_GAME:
		sg := StartGameDTO{
			WhiteId: r.WhitePlayer.User.Id.String(),
			BlackId: r.BlackPlayer.User.Id.String(),
		}
		p, err := json.Marshal(sg)
		if err != nil {
			return
		}
		e.Payload = p

	default:
		slog.Warn("event had unknown action", fn, "action", action)
		return
	}

	e.Action = action
	r.WhitePlayer.writeEventBuffer <- e
	r.BlackPlayer.writeEventBuffer <- e
}
