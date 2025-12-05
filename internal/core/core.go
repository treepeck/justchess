// TODO: bunch of unsafe code. Any type should never be used.
package core

import (
	"encoding/json"
	"justchess/internal/db"
	"log"

	"github.com/rabbitmq/amqp091-go"

	"github.com/treepeck/chego"
	"github.com/treepeck/gatekeeper/pkg/event"
	"github.com/treepeck/gatekeeper/pkg/mq"
)

/*
Core is responsible for handling incoming events from both active rooms and
the Gatekeeper server.
*/
type Core struct {
	mm   matchmaking
	repo db.Repo
	// AMQP channel to communicate with Gatekeeper.
	channel *amqp091.Channel
	// Active game rooms which will process player's moves.
	rooms map[string]*room
	// Buffered channel to queue up players' requests and allow multiple rooms
	// to handle them concurrently.
	request chan event.Internal
	// Add game room.
	add chan addRoomRes
	// Remove game room.
	remove chan string
	// Unbuffered channel of encoded rooms' responses to avoid deadlocks.
	response chan []byte
}

func NewCore(ch *amqp091.Channel, r db.Repo) Core {
	c := Core{
		repo:     r,
		channel:  ch,
		rooms:    make(map[string]*room),
		request:  make(chan event.Internal, 192),
		add:      make(chan addRoomRes),
		remove:   make(chan string),
		response: make(chan []byte),
	}

	c.mm = newMatchmaking(c.add)
	go c.mm.handleEvents()

	return c
}

/*
EventBus consumes events from the "gate" queue and internal [Core] channels.
Each event is forwarded to the corresponding handler after consuming.

Designed to run as a separate goroutine until the program shuts down.  Panics if
the "gate" queue cannot be consumed.  Implements concurrent worker pool pattern.
*/
func (c *Core) EventBus() {
	go mq.Consume(c.channel, "gate", c.request)

	// Listen for room responses in a separate goroutine to avoid deadlocks.
	go func() {
		for {
			select {
			case res := <-c.add:
				r := newRoom(res.roomId, res.players[0], res.players[1],
					c.response, c.remove, res.timeControl, res.timeBonus)
				go r.handleEvents()
				c.rooms[res.roomId] = r
				log.Printf("room %s added", res.roomId)
				c.publishEncoded(event.ActionAddRoom, res.players, res.roomId)

			case roomId := <-c.remove:
				delete(c.rooms, roomId)
				log.Printf("room %s removed", roomId)
				c.publishEncoded(event.ActionRemoveRoom, nil, roomId)

			case raw := <-c.response:
				mq.Publish(c.channel, "core", raw)
			}
		}
	}()

	// Endless loop to queue up players' requests.
	for {
		c.handleRequest(<-c.request)
	}
}

/*
handleRequest forwards event to the corresponding room if that room is ready
for processing (its channel is empty).
*/
func (c *Core) handleRequest(e event.Internal) {
	r, exists := c.rooms[e.RoomId]
	if e.RoomId != "hub" && !exists {
		log.Printf("event to room %s which doesn't exist from %s", e.RoomId, e.ClientId)
		return
	}

	switch e.Action {
	case event.ActionJoinMatchmaking:
		var p roomParams
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			log.Printf("malformed request from %s: %s", e.ClientId, err)
			// TODO: ban players who send many malformed requests.
			return
		}
		c.mm.join <- joinMatchmakingReq{
			playerId: e.ClientId,
			params:   p,
		}

	case event.ActionLeaveMatchmaking:
		c.mm.leave <- e.ClientId

	case event.ActionJoinRoom:
		r.join <- e.ClientId

	case event.ActionLeaveRoom:
		r.leave <- e.ClientId

	case event.ActionMakeMove:
		var m chego.Move
		if err := json.Unmarshal(e.Payload, &m); err != nil {
			log.Printf("malformed request from %s: %s", e.ClientId, err)
			// TODO: handle malformed event function.
			return
		}
		r.move <- moveReq{playerId: e.ClientId, move: m}
	}
}

/*
publishEncoded encodes the specified internal event and calls [mq.Publish].
*/
func (c *Core) publishEncoded(a event.Action, p any, roomId string) {
	raw, err := event.EncodeInternal(a, p, "", roomId)
	if err != nil {
		log.Printf("cannot encode internal event: %s", err)
		return
	}

	mq.Publish(c.channel, "core", raw)
}

func (c *Core) handleMalformedEvent() {
}
