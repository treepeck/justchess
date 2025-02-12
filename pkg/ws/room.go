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
	// After making the move the client send it to this channel.
	moves chan []byte
}

func newRoom(control, bonus uint8) *room {
	return &room{
		id:         uuid.New(),
		clients:    make(map[*client]struct{}),
		game:       game.NewGame(enums.Unknown, nil, control, bonus),
		register:   make(chan *client),
		unregister: make(chan *client),
		moves:      make(chan []byte),
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

		case m := <-r.moves:
			r.handleMove(m)
		}
	}
}

func (r *room) addClient(c *client) {
	// Skip already connected clients.
	for connectedC := range r.clients {
		if connectedC.id == c.id {
			return
		}
	}

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

func (r *room) handleMove(msg []byte) {
	move := bitboard.NewMove(int(msg[0]), int(msg[1]), enums.MoveType(msg[2]))
	if r.game.ProcessMove(move) {
		r.broadcastLastMove()
	}
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

func (r *room) broadcastLastMove() {
	if len(r.game.Moves) < 1 {
		return
	}
	lastMove := r.game.Moves[len(r.game.Moves)-1]
	// The LAST_MOVE message consists of 4 parts:
	//   1 - SAN of the completed move;
	//   2 - FEN of the current board state;
	//   3 - Legal moves for the next player.
	//   4 - Message type: LAST_MOVE.
	// First 3 parts of the message are separated by a 0xFF byte.
	msg := make([]byte, 0)
	msg = append(msg, []byte(lastMove.SAN)...)
	msg = append(msg, 0xFF) // Separator.
	msg = append(msg, []byte(lastMove.FEN)...)
	msg = append(msg, 0xFF) // Separator.
	for _, move := range r.game.Bitboard.LegalMoves {
		msg = append(msg, byte(move.To()), byte(move.From()), byte(move.Type()))
	}
	msg = append(msg, LAST_MOVE)
	for c := range r.clients {
		c.send <- msg
	}
}
