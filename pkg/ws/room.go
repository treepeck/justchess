package ws

import (
	"encoding/json"
	"justchess/pkg/auth"
	"justchess/pkg/game"
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"
	"log"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
)

type RoomStatus int

const (
	// Waiting for at least 2 clients to start a game.
	OPEN RoomStatus = iota
	IN_PROGRESS
	WHITE_DISCONNECTED
	BLACK_DISCONNECTED
	// Room denies all incomming requests and waits until all clients leaves.
	OVER
)

// Room is a middleman between a particular game.Game instance and clients.
// After the game is over or all clients are disconnected, the room removes itself from the Hub.
//
// Room accepts new connections in each status, except OVER.
// Room methods cannot be called outside the the r.handleRoutine!
type Room struct {
	// The room must be able to remove itself from the Hub.
	hub        *Hub
	status     RoomStatus
	game       *game.Game
	creatorId  uuid.UUID
	register   chan *client
	unregister chan *client
	move       chan bitboard.Move
	chat       chan string
	// When one of the players runs out of time, the game sends the msg to the timeout
	// channel, so that the room can notify the players about the game result.
	timeout chan struct{}
	clients map[*client]struct{}
}

func newRoom(h *Hub, id uuid.UUID, control, bonus int) *Room {
	return &Room{
		hub:        h,
		status:     OPEN,
		creatorId:  id,
		clients:    make(map[*client]struct{}),
		register:   make(chan *client),
		unregister: make(chan *client),
		move:       make(chan bitboard.Move),
		chat:       make(chan string),
		timeout:    make(chan struct{}),
		game:       game.NewGame(nil, control*60, bonus),
	}
}

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

	go c.readRoutine()
	go c.writeRoutine()

	r.register <- c
}

// handleRoutine handles incomming connections, disconnections and completed moves.
// Hub is responsible for terminating this routine.
func (r *Room) handleRoutine() {
	for {
		select {
		case c, ok := <-r.register:
			if !ok {
				return
			}
			r.add(c)

		case c := <-r.unregister:
			r.remove(c)

		case m := <-r.move:
			r.handle(m)

		case msg := <-r.chat:
			r.broadcastChat(msg)

		case <-r.timeout:
			r.endGame()
		}
	}
}

// add denies multiple connections from a single peer.
func (r *Room) add(c *client) {
	for connected := range r.clients {
		if connected.id == c.id || r.status == OVER {
			close(c.send)
			return
		}
	}

	r.clients[c] = struct{}{}
	log.Printf("client %s registered\n", c.id.String())

	switch r.status {
	case OPEN:
		// Start the game.
		if len(r.clients) == 2 {
			r.startGame()
		}

	case WHITE_DISCONNECTED:
		if r.game.WhiteId == c.id {
			r.status = IN_PROGRESS
		}

	case BLACK_DISCONNECTED:
		if r.game.BlackId == c.id {
			r.status = IN_PROGRESS
		}
	}

	// Notify the client about all completed moves.
	for _, m := range r.game.Moves {
		c.send <- r.serialize(LAST_MOVE, r.formatLastMove(m))
	}
	r.broadcast(r.serialize(ROOM_STATUS, r.formatRoomStatus()))
}

// remove terminates the game.DecrementTime routine.
func (r *Room) remove(c *client) {
	if _, ok := r.clients[c]; !ok {
		return
	}

	close(c.send)
	delete(r.clients, c)

	log.Printf("client %s unregistered\n", c.id.String())

	if len(r.clients) == 0 {
		if r.status != OPEN && r.status != OVER {
			r.game.End <- struct{}{}
		}

		r.hub.remove(r)
	}
}

func (r *Room) startGame() {
	r.status = IN_PROGRESS
	go r.game.DecrementTime(r.timeout)

	players := [2]uuid.UUID{}
	i := 0
	for c := range r.clients {
		players[i] = c.id
		i++
	}

	// Randomly select player`s sides.
	if rand.Intn(2) == 1 {
		r.game.WhiteId = players[0]
		r.game.BlackId = players[1]
	} else {
		r.game.WhiteId = players[1]
		r.game.BlackId = players[0]
	}
}

func (r *Room) handle(m bitboard.Move) {
	if r.status == OPEN || r.status == OVER {
		return
	}

	if r.game.ProcessMove(m) {
		lastMove := r.game.Moves[len(r.game.Moves)-1]
		r.broadcast(r.serialize(LAST_MOVE, r.formatLastMove(lastMove)))

		if r.game.Result != enums.Unknown {
			r.endGame()
		}
	}
}

func (r *Room) broadcastChat(message string) {
	data, err := json.Marshal(ChatData{Message: message})
	if err != nil {
		log.Printf("cannot Marshal message: %v\n", err)
		return
	}

	msg, _ := json.Marshal(Message{Type: CHAT, Data: data})

	for c := range r.clients {
		c.send <- msg
	}
}

// endGame ends the game, broadcasts room status and game result.
func (r *Room) endGame() {
	r.status = OVER
	r.broadcast(r.serialize(ROOM_STATUS, r.formatRoomStatus()))

	// Broadcast game result.
	data, _ := json.Marshal(GameResultData{Result: r.game.Result})
	msg, _ := json.Marshal(Message{Type: GAME_RESULT, Data: data})
	r.broadcast(msg)
}

func (r *Room) formatRoomStatus() RoomStatusData {
	return RoomStatusData{
		Status:    r.status,
		WhiteId:   r.game.WhiteId.String(),
		BlackId:   r.game.BlackId.String(),
		WhiteTime: r.game.WhiteTime,
		BlackTime: r.game.BlackTime,
		Clients:   len(r.clients),
	}
}

func (r *Room) formatLastMove(move game.CompletedMove) LastMoveData {
	legalMoves := make([]MoveData, len(r.game.Bitboard.LegalMoves))
	for i, m := range r.game.Bitboard.LegalMoves {
		legalMoves[i] = MoveData{To: m.To(), From: m.From(), Type: m.Type()}
	}

	return LastMoveData{
		SAN:        move.SAN,
		FEN:        move.FEN,
		TimeLeft:   move.TimeLeft,
		LegalMoves: legalMoves,
	}
}

// serialize can recieve data only of the specified types!
func (r *Room) serialize(mt MessageType, data any) []byte {
	raw, err := json.Marshal(data)
	if err != nil {
		log.Printf("cannot Marshal data: %v\n", err)
		return nil
	}

	msg, err := json.Marshal(Message{Type: mt, Data: raw})
	if err != nil {
		log.Printf("cannot Marshal message: %v\n", err)
		return nil
	}
	return msg
}

func (r *Room) broadcast(msg []byte) {
	for c := range r.clients {
		c.send <- msg
	}
}
