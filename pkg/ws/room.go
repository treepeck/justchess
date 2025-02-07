package ws

import (
	"justchess/pkg/game"
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"log"
	"math/rand"

	"github.com/google/uuid"
)

type room struct {
	id uuid.UUID
	// Use of empty struct{} is an optimization.
	clients    map[*client]struct{}
	game       *game.Game
	register   chan *client
	unregister chan *client
}

func newRoom(control, bonus uint8) *room {
	bb := bitboard.NewBitboard([12]uint64{0xFF00, 0xFF000000000000, 0x42,
		0x4200000000000000, 0x24, 0x2400000000000000, 0x7E, 0x8100000000000000,
		0x8, 0x800000000000000, 0x10, 0x1000000000000000}, enums.White,
		[4]bool{true, true, true, true}, -1, 0, 0)
	return &room{
		id:         uuid.New(),
		clients:    make(map[*client]struct{}),
		game:       game.NewGame(enums.Unknown, bb, control, bonus),
		register:   make(chan *client),
		unregister: make(chan *client),
	}
}

func (r *room) run() {
	for {
		select {
		case c, ok := <-r.register:
			if !ok {
				return
			}
			r.addClient(c)
			r.startGame()

		case c := <-r.unregister:
			r.removeClient(c)
		}
	}
}

func (r *room) addClient(c *client) {
	if len(r.clients) < 2 {
		r.clients[c] = struct{}{}
		c.currentRoom = r
		// Redirect the client to the room.
		msg := make([]byte, 17)
		copy(msg[0:16], r.id[:])
		msg[16] = REDIRECT
		c.send <- msg
		log.Printf("client %s added\n", c.id.String())
	}
}

func (r *room) removeClient(c *client) {
	delete(r.clients, c)
	log.Printf("client %s removed\n", c.id.String())
	// If there are no clients left in the room, remove it.
	if len(r.clients) == 0 {
		close(r.register)
		// TRICK: remove the room from the manager.
		c.manager.remove <- r
	}
}

func (r *room) startGame() {
	if len(r.clients) != 2 {
		return
	}
	// Randomly generate player`s colors.
	for c := range r.clients {
		if r.game.WhiteId == uuid.Nil {
			r.game.WhiteId = c.id
		} else {
			r.game.BlackId = c.id
		}
	}
	if rand.Intn(2) == 1 {
		tmp := r.game.BlackId
		r.game.BlackId = r.game.WhiteId
		r.game.BlackId = tmp
	}
	go r.game.DecrementTime()
	// Noify the clients that the game has begun.
	r.broadcastGameInfo()
}

func (r *room) broadcastGameInfo() {
	msg := make([]byte, 36)
	copy(msg[0:16], r.game.WhiteId[:])
	copy(msg[16:32], r.game.BlackId[:])
	msg[32] = byte(r.game.Result)
	msg[33] = r.game.TimeControl
	msg[34] = r.game.TimeBonus
	msg[35] = GAME_INFO
	for c := range r.clients {
		c.send <- msg
	}
}
