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

type CoreService struct {
	rooms       map[string]*room
	matchmaking map[waitRoom]struct{}
	EventBus    chan types.MetaEvent
	pool        *sql.DB
	channel     *amqp091.Channel
}

/*
Core is responsible for handling incoming events from both active rooms and
the Gatekeeper.
*/
type Core struct {
	rooms       map[string]*room
	matchmaking map[waitRoom]struct{}
	EventBus    chan types.MetaEvent
	pool        *sql.DB
	channel     *amqp091.Channel
}

func NewCore(ch *amqp091.Channel, pool *sql.DB) *Core {
	return &Core{
		rooms:       make(map[string]*room),
		matchmaking: make(map[waitRoom]struct{}),
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
				// Remove the wait room if the creator disconnects.
				for waitRoom := range c.matchmaking {
					if waitRoom.creatorId == e.ClientId {
						delete(c.matchmaking, waitRoom)
					}
				}
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
It's a Gatekeeper's responsibility to ensure that the client isn't already in the
room [mark clients which have already entered matchmaking].
*/
func (c *Core) handleEnterMatchmaking(e types.MetaEvent) {
	// Deny the request if the payload is malformed.
	var dto types.EnterMatchmaking
	if json.Unmarshal(e.Payload, &dto) != nil {
		return
	}

	// Search for the match.
	for waitRoom := range c.matchmaking {
		if waitRoom.timeControl != dto.TimeControl ||
			waitRoom.timeBonus != dto.TimeBonus {
			continue
		}

		// Delete the wait room from matchmaking after the match was found.
		delete(c.matchmaking, waitRoom)

		// Create new game room.
		id := randgen.GenId(randgen.IdLen)
		r := newRoom(
			id,
			waitRoom.creatorId, // Player 1.
			e.ClientId,         // Player 2.
			c.EventBus,
			waitRoom.timeControl,
			waitRoom.timeBonus,
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
		return
	}

	// If there isn't room with the same parameters, create a new one.
	waitRoom := waitRoom{
		creatorId:   e.ClientId,
		timeControl: dto.TimeControl,
		timeBonus:   dto.TimeBonus,
	}

	// Do not notify clients here, just display that the game is searching on a
	// frontend.
	c.matchmaking[waitRoom] = struct{}{}
}
