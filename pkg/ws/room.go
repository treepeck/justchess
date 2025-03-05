package ws

import (
	"justchess/pkg/game"
	"justchess/pkg/game/bitboard"
	"log"

	"github.com/google/uuid"
)

type Room struct {
	hub        *Hub
	creatorId  uuid.UUID
	clients    map[*client]struct{}
	register   chan *client
	unregister chan *client
	move       chan bitboard.Move
	game       *game.Game
}

func newRoom(h *Hub, id uuid.UUID, control, bonus int) *Room {
	return &Room{
		hub:        h,
		creatorId:  id,
		clients:    make(map[*client]struct{}),
		register:   make(chan *client),
		unregister: make(chan *client),
		move:       make(chan bitboard.Move),
		game:       game.NewGame(nil, control, bonus),
	}
}

// handleRoutine handles incomming connections, disconnections and completed moves.
// Hub is responsible for terminating this routine.
func (r *Room) handleRoutine() {
	for {
		select {
		case c := <-r.register:
			r.add(c)

		case c := <-r.unregister:
			r.remove(c)

		case m := <-r.move:
			r.handle(m)
		}
	}
}

func (r *Room) add(c *client) {
	log.Printf("incomming connection: %s\n", c.id.String())
}

func (r *Room) remove(c *client) {
	log.Printf("breaking connection: %s\n", c.id.String())
}

func (r *Room) handle(m bitboard.Move) {
	log.Printf("incomming move: %v\n", m)
}
