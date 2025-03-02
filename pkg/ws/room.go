package ws

import (
	"justchess/pkg/auth"
	"justchess/pkg/game"
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"log"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
)

type RoomStatus byte

const (
	// Noone connected.
	ABANDONED RoomStatus = iota
	// One player connected but the game was not started.
	WAITING
	// Both players connected - the game is in progress.
	IN_PROGRESS
	// White player disconnected - game paused.
	WHITE_DISCONNECTED
	// Black player disconnected - game paused.
	BLACK_DISCONNECTED
	// The game is over, but the both players are connected.
	OVER
)

type Room struct {
	creatorId  uuid.UUID
	game       *game.Game
	status     RoomStatus
	register   chan *client
	unregister chan *client
	move       chan bitboard.Move
	clients    map[*client]struct{}
	// When one of the players runs out of time, the game sends the msg to the timeout
	// channel, so that the room can notify the players about the game result.
	timeout chan struct{}
}

func NewRoom(creatorId uuid.UUID, timeControl, timeBonus byte) *Room {
	return &Room{
		creatorId:  creatorId,
		game:       game.NewGame(nil, timeControl, timeBonus),
		status:     ABANDONED,
		register:   make(chan *client),
		unregister: make(chan *client),
		move:       make(chan bitboard.Move),
		clients:    make(map[*client]struct{}),
		timeout:    make(chan struct{}),
	}
}

// HandleNewConnection creates a new client and registers it in the Room.
// To be connected, client must provide a valid access JWT as a GET request param.
func (r *Room) HandleNewConnection(rw http.ResponseWriter, req *http.Request) {
	encoded := req.URL.Query().Get("access")
	access, err := auth.DecodeToken(encoded, 1)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	idStr, err := access.Claims.GetSubject()
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	c := newClient(id, conn)
	c.room = r

	go c.readPump()
	go c.writePump()

	r.register <- c
}

func (r *Room) EventPump() {
	for {
		select {
		case c, ok := <-r.register:
			if !ok {
				return
			}
			r.registerClient(c)
			r.broadcastRoomInfo()

		case c := <-r.unregister:
			if _, ok := r.clients[c]; ok {
				r.unregisterClient(c)
				r.broadcastRoomInfo()
			}

		case m := <-r.move:
			r.handleMove(m)

		case <-r.timeout:
			r.status = OVER
			r.broadcastRoomInfo()
			r.broadcastGameResult()
		}
	}
}

func (r *Room) registerClient(c *client) {
	for connected := range r.clients {
		// Deny multiple connections from a single peer.
		if connected.id == c.id {
			close(c.send)
			return
		}
	}

	switch r.status {
	case IN_PROGRESS, OVER:
		// Deny any connections.
		return

	case ABANDONED:
		r.clients[c] = struct{}{}
		r.status = WAITING

	case WAITING:
		r.clients[c] = struct{}{}
		r.startGame()

	case WHITE_DISCONNECTED:
		if c.id != r.game.WhiteId {
			return
		}
		r.clients[c] = struct{}{}
		r.sendGame(c)
		r.status = IN_PROGRESS

	case BLACK_DISCONNECTED:
		if c.id != r.game.BlackId {
			return
		}
		r.clients[c] = struct{}{}
		r.sendGame(c)
		r.status = IN_PROGRESS
	}
	log.Printf("client %s added\n", c.id.String())
}

func (r *Room) unregisterClient(c *client) {
	close(c.send)
	delete(r.clients, c)
	log.Printf("client %s removed\n", c.id.String())

	switch r.status {
	case WAITING, WHITE_DISCONNECTED, BLACK_DISCONNECTED:
		r.status = OVER

		// Terminate the DecrementTime goroutine.
		r.game.End <- struct{}{}
		log.Printf("game ended\n")

	case IN_PROGRESS:
		if c.id == r.game.WhiteId {
			r.status = WHITE_DISCONNECTED
		} else {
			r.status = BLACK_DISCONNECTED
		}
	}
}

func (r *Room) startGame() {
	r.status = IN_PROGRESS

	playerIDs := make([]uuid.UUID, 0)
	for c := range r.clients {
		playerIDs = append(playerIDs, c.id)
	}

	// Randomly select player`s sides.
	if rand.Intn(2) == 1 {
		r.game.WhiteId = playerIDs[0]
		r.game.BlackId = playerIDs[1]
	} else {
		r.game.WhiteId = playerIDs[1]
		r.game.BlackId = playerIDs[0]
	}

	go r.game.DecrementTime(r.timeout)
	log.Printf("game has been started")
}

// handleMoves rejects the move if the game has not been started yet or
// one of the players is disconnected.
func (r *Room) handleMove(m bitboard.Move) {
	if r.status == ABANDONED || r.status == WAITING || r.status == OVER {
		return
	}

	// If the move is legal, broadcast it.
	if r.game.ProcessMove(m) {
		r.broadcastLastMove(r.game.Moves[len(r.game.Moves)-1])

		if r.game.Result != enums.Unknown {
			r.broadcastGameResult()
		}
	}
}

func (r *Room) broadcast(msg []byte) {
	for c := range r.clients {
		c.send <- msg
	}
}

// broadcastRoomInfo broadcast room status and connected client`s id if the
// game is in progress.
func (r *Room) broadcastRoomInfo() {
	msg := []byte{byte(r.status)}

	if r.status == IN_PROGRESS {
		msg = append(msg, r.game.WhiteId[:]...)
		msg = append(msg, r.game.BlackId[:]...)
	}

	msg = append(msg, ROOM_INFO)
	r.broadcast(msg)
}

// broadcastLastMove broadcasts the completed move and current legal moves.
func (r *Room) broadcastLastMove(move game.CompletedMove) {
	msg := make([]byte, 0)
	msg = append(msg, encodeCompletedMove(move)...)

	// The 0xAF byte separates the completed moves from current legal moves.
	msg = append(msg, 0xAF)
	for _, m := range r.game.Bitboard.LegalMoves {
		msg = append(msg, byte(m.To()), byte(m.From()), byte(m.Type()))
	}
	msg = append(msg, LAST_MOVE)

	r.broadcast(msg)
}

func (r *Room) broadcastGameResult() {
	msg := []byte{byte(r.game.Result), RESULT}
	r.broadcast(msg)
}

// sendGame sends all completed moves, current legal moves and players remaining time.
// This information allows the user to recover game state after reconnection.
func (r *Room) sendGame(to *client) {
	if len(r.game.Moves) == 0 {
		return
	}

	msg := make([]byte, 0)
	for _, move := range r.game.Moves {
		msg = append(msg, encodeCompletedMove(move)...)
	}

	// The 0xAF byte separates the completed moves from current legal moves.
	msg = append(msg, 0xAF)
	for _, move := range r.game.Bitboard.LegalMoves {
		msg = append(msg, byte(move.To()), byte(move.From()), byte(move.Type()))
	}

	msg = append(msg, []byte{r.game.WhiteTime, r.game.BlackTime, GAME}...)
	to.send <- msg
}

// encodeCompletedMove encodes the move into 3 parts:
//  1. SAN of the completed move;
//  2. FEN of the current board state;
//  3. Time left after completing the move.
//
// First and second parts vary in length and are separated by the 0xFF byte.
func encodeCompletedMove(move game.CompletedMove) (decoded []byte) {
	decoded = append(decoded, []byte(move.SAN)...)
	decoded = append(decoded, 0xFF) // Separator.
	decoded = append(decoded, []byte(move.FEN)...)
	decoded = append(decoded, 0xFF) // Separator.
	return append(decoded, move.TimeLeft)
}
