package ws

import (
	"chess-api/models"
	"chess-api/models/enums"
	"chess-api/models/helpers"
	"encoding/json"
	"log/slog"
	"math/rand"
	"sync"

	"github.com/google/uuid"
)

type Room struct {
	sync.Mutex
	Id          uuid.UUID    `json:"id"`
	Game        *models.Game `json:"game"`
	Owner       models.User  `json:"owner"`
	WhitePlayer *Client
	BlackPlayer *Client
}

type CreateRoomDTO struct {
	Control enums.Control `json:"control"`
	Bonus   uint          `json:"bonus"`
	Owner   models.User   `json:"owner"`
}

type GetRoomDTO struct {
	Control string `json:"control"`
	Bonus   string `json:"bonus"`
}

func NewRoom(cr CreateRoomDTO, owner *Client) *Room {
	r := &Room{
		Id:          uuid.New(),
		Owner:       cr.Owner,
		Game:        nil,
		WhitePlayer: nil,
		BlackPlayer: nil,
	}

	// randomize side selection
	whiteId := uuid.Nil
	blackId := uuid.Nil
	if rand.Intn(2) == 1 {
		r.WhitePlayer = owner
		whiteId = owner.User.Id
	} else {
		r.BlackPlayer = owner
		blackId = owner.User.Id
	}
	owner.changeRoom(r.Id)

	r.Game = models.NewGame(r.Id, cr.Control, cr.Bonus, whiteId, blackId)
	return r
}

// Adds client to the room if:
// 1. The game hasn`t been started yet and there is availible side.
// 2. The game has been started, but the client just reconnects.
func (r *Room) AddPlayer(c *Client) {
	r.Lock()
	defer r.Unlock()

	if r.Game.Status == enums.Waiting {
		// take the availible side
		if r.Game.WhiteId == uuid.Nil {
			r.WhitePlayer = c
			r.Game.WhiteId = c.User.Id
			c.changeRoom(r.Id)
		} else if r.Game.BlackId == uuid.Nil {
			r.BlackPlayer = c
			r.Game.BlackId = c.User.Id
			c.changeRoom(r.Id)
		}
		r.Game.Status = enums.Continues
	} else {
		// handle reconnection
		if r.Game.WhiteId == c.User.Id {
			r.WhitePlayer = c
			c.changeRoom(r.Id)
		} else if r.Game.BlackId == c.User.Id {
			r.BlackPlayer = c
			c.changeRoom(r.Id)
		}
	}

	// if the user joined, update a frontend game state
	if c.currentRoomId != uuid.Nil {
		r.Broadcast(UPDATE_GAME)
	}
}

func (r *Room) Broadcast(action string) {
	fn := slog.String("func", "room.broadcast")

	var e Event
	switch action {
	case UPDATE_GAME:
		p, err := json.Marshal(r.Game)
		if err != nil {
			slog.Warn("cannot Marshal game", fn, "err", err)
			return
		}
		e.Payload = p

	default:
		slog.Warn("event had unknown action", fn, "action", action)
		return
	}

	e.Action = action
	if r.WhitePlayer != nil {
		r.WhitePlayer.writeEventBuffer <- e
	}
	if r.BlackPlayer != nil {
		r.BlackPlayer.writeEventBuffer <- e
	}
}

func (r *Room) HandleTakeMove(move helpers.MoveDTO, c *Client) {
	index := r.Game.Moves.Depth() + 1
	isEven := index%2 == 0
	// for the white player the current move number must be odd
	if !isEven && c.User.Id == r.Game.WhiteId {
		if r.Game.TakeMove(move) {
			r.Broadcast(UPDATE_GAME)
		}
	} else if isEven && c.User.Id == r.Game.BlackId {
		if r.Game.TakeMove(move) {
			r.Broadcast(UPDATE_GAME)
		}
	}
}

func (r *Room) HandleGetGame(c *Client) {
	fn := slog.String("func", "HandleGetName")
	// send updated game info back to the client
	p, err := json.Marshal(r.Game)
	if err != nil {
		slog.Warn("cannot Marshal game", fn, "err", err)
		return
	}
	e := Event{
		Action:  UPDATE_GAME,
		Payload: p,
	}
	c.writeEventBuffer <- e
}
