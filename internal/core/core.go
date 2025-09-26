package core

import (
	"database/sql"
	"encoding/json"
	"log"
	"strconv"

	"github.com/rabbitmq/amqp091-go"

	"justchess/internal/randgen"

	"github.com/treepeck/chego"
	"github.com/treepeck/gatekeeper/pkg/mq"
	"github.com/treepeck/gatekeeper/pkg/types"
)

/*
Core is responsible for handling incoming events from both active rooms and
the Gatekeeper.
*/
type Core struct {
	matchmaking *matchmaking
	// Active game rooms.
	rooms    map[string]*room
	EventBus chan types.MetaEvent
	pool     *sql.DB
	channel  *amqp091.Channel
}

func NewCore(ch *amqp091.Channel, pool *sql.DB) *Core {
	return &Core{
		matchmaking: newMatchmaking(),
		rooms:       make(map[string]*room),
		EventBus:    make(chan types.MetaEvent),
		pool:        pool,
		channel:     ch,
	}
}

/*
Run consequentially (one at a time) accepts events from the Bus and routes
them to the corresponding handler function.
*/
func (c *Core) Run() {
	for {
		e := <-c.EventBus

		switch e.Action {
		// Client events.

		case types.ActionEnterMatchmaking:
			c.handleEnterMatchmaking(e)

		// Forward the incomming move to the existing game room which will
		// handle it.
		case types.ActionMakeMove:
			// Validate and decode.
			if r, exists := c.rooms[e.RoomId]; exists {
				if p, err := strconv.Atoi(string(e.Payload)); err == nil {
					r.move <- moveDTO{
						playerId: e.ClientId,
						move:     chego.Move(p),
					}
				}
			}

		case types.ActionJoinRoom:
			if r, exists := c.rooms[e.RoomId]; exists {
				r.handleJoin(e.ClientId)
			}

		case types.ActionLeaveRoom:
			if r, exists := c.rooms[e.RoomId]; exists {
				r.handleLeave(e.ClientId)
			} else {
				c.matchmaking.leave(e.ClientId)
			}

		// Room events.
		case types.ActionCompletedMove:
			if _, exists := c.rooms[e.RoomId]; exists {
				if raw, err := json.Marshal(e); err == nil {
					mq.Publish(c.channel, "core", raw)
				}
			}

		case types.ActionRemoveRoom:
			if _, exists := c.rooms[e.RoomId]; exists {
				if raw, err := json.Marshal(e); err == nil {
					// Remove the room.
					delete(c.rooms, e.RoomId)

					// Notify the gatekeeper about removed room.
					mq.Publish(c.channel, "core", raw)
				}
			}
		}
	}
}

/*
handleEnterMatchmaking denies the request if the client is already in game or
matchmaking room.
*/
func (c *Core) handleEnterMatchmaking(e types.MetaEvent) {
	// Deny the request if the player has already entered matchmaking.
	if c.matchmaking.hasEntered(e.ClientId) {
		return
	}

	// Deny the request if the player is playing the game.
	for _, r := range c.rooms {
		if e.ClientId == r.whiteId || e.ClientId == r.blackId {
			return
		}
	}

	// Deny the request if the payload is malformed.
	var dto types.EnterMatchmaking
	if json.Unmarshal(e.Payload, &dto) != nil {
		return
	}

	// Search for the match.
	matchId := c.matchmaking.match(dto)
	// If there isn't room with the same parameters, create a new one.
	// Do not notify clients here, just display that the game is searching on a
	// frontend.
	if matchId == "" {
		c.matchmaking.enter(e.ClientId, dto)
		return
	}

	c.matchmaking.leave(matchId)

	// Create new game room.
	id := randgen.GenId(randgen.IdLen)
	r := newRoom(
		id,
		matchId,    // Player 1.
		e.ClientId, // Player 2.
		c.EventBus,
		dto.TimeControl,
		dto.TimeBonus,
	)

	go r.run(id)

	// Add room to the core.
	c.rooms[id] = r

	// Notify the clients about start of the game.
	p, err := json.Marshal(types.AddRoom{
		WhiteId: r.whiteId,
		BlackId: r.blackId,
	})
	if err != nil {
		log.Printf("cannot encode add room payload: %s", err)
		return
	}

	raw, err := json.Marshal(types.MetaEvent{
		Action:  types.ActionAddRoom,
		Payload: p,
		RoomId:  id,
	})
	if err != nil {
		log.Printf("cannot encode add room event: %s", err)
		return
	}
	mq.Publish(c.channel, "core", raw)
}
