package core

import (
	"encoding/json"
	"justchess/internal/db"
	"justchess/internal/randgen"
	"log"

	"github.com/rabbitmq/amqp091-go"

	"github.com/treepeck/gatekeeper/pkg/event"
	"github.com/treepeck/gatekeeper/pkg/mq"
)

/*
Core is responsible for handling incoming events from both active rooms and
the Gatekeeper.
*/
type Core struct {
	mm matchmaking
	// Active game rooms.
	rooms    map[string]*room
	EventBus chan event.Internal
	repo     *db.Repo
	channel  *amqp091.Channel
}

func NewCore(ch *amqp091.Channel, r *db.Repo) *Core {
	return &Core{
		mm:       make(matchmaking),
		rooms:    make(map[string]*room),
		EventBus: make(chan event.Internal),
		repo:     r,
		channel:  ch,
	}
}

/*
Run consequentially (one at a time) accepts events from the EventBus channel and
routes them to the corresponding handler function.
*/
func (c *Core) Run() {
	for {
		e := <-c.EventBus

		switch e.Action {
		case event.ActionEnterMatchmaking:
			c.handleEnterMatchmaking(e)

		case event.ActionLeaveRoom:
			// c.handle
		}
	}
}

/*
handleEnterMatchmaking denies the request if the client is already in game or
matchmaking room.
*/
func (c *Core) handleEnterMatchmaking(e event.Internal) {
	var dto matchmakingDTO
	if err := json.Unmarshal(e.Payload, &dto); err != nil {
		log.Printf("cannot decode event payload: %s", err)
		return
	}

	// Deny the request if the player has already entered matchmaking or game.
	if c.mm.hasEntered(e.SenderId) {
		return
	} else {
		for _, r := range c.rooms {
			if e.SenderId == r.whiteId || e.SenderId == r.blackId {
				return
			}
		}
	}

	// Search for the match.
	matchId := c.mm.match(dto)
	// If there isn't room with the same parameters, create a new one.
	// Don't notify clients here, just display that the game is searching on a
	// frontend.
	if matchId == "" {
		c.mm.enter(e.SenderId, dto)
		return
	}

	c.mm.leave(matchId)

	// Create new game room.
	roomId := randgen.GenId(randgen.IdLen)
	r := newRoom(
		roomId,
		matchId,    // Player 1.
		e.SenderId, // Player 2.
		c.EventBus,
		dto.TimeControl,
		dto.TimeBonus,
	)

	go r.run()

	// Store active game room.
	c.rooms[roomId] = r

	// Redirect players to the room.
	if raw, err := json.Marshal([2]string{r.whiteId, r.blackId}); err == nil {
		mq.Publish(c.channel, "core", event.Internal{
			Action:   event.ActionAddRoom,
			Payload:  raw,
			SenderId: roomId,
		})
	}
}
