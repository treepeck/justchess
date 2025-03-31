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
	"sync"

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
type Room struct {
	sync.Mutex
	// The room must be able to remove itself from the Hub.
	id          uuid.UUID
	creatorName string
	hub         *Hub
	isVSEngine  bool
	status      RoomStatus
	game        *game.Game
	clients     map[*client]struct{}
}

func newRoom(h *Hub, creatorName string, isVSEngine bool, control, bonus int) *Room {
	return &Room{
		id:          uuid.New(),
		creatorName: creatorName,
		hub:         h,
		isVSEngine:  isVSEngine,
		status:      OPEN,
		clients:     make(map[*client]struct{}),
		game:        game.NewGame(nil, control*60, bonus),
	}
}

func (r *Room) HandleNewConnection(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context().Value(auth.Cms)
	if ctx == nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	cms := ctx.(auth.Claims)

	// Guest users cannot play with other users, only vs engine.
	if !r.isVSEngine && cms.Role == auth.RoleGuest {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	c := newClient(cms.Id, cms.Name, cms.Role == auth.RoleGuest, conn)
	c.room = r

	go c.readRoutine()
	go c.writeRoutine()

	r.register(c)
}

// register denies multiple connections from a single peer.
func (r *Room) register(c *client) {
	r.Lock()
	defer r.Unlock()

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
		if r.isVSEngine || len(r.clients) == 2 {
			r.startGame()
		}

	case WHITE_DISCONNECTED:
		if r.game.White == c.name {
			r.status = IN_PROGRESS
		}

	case BLACK_DISCONNECTED:
		if r.game.Black == c.name {
			r.status = IN_PROGRESS
		}
	}

	// Notify the client about all completed moves.
	for _, m := range r.game.Moves {
		c.send <- r.serialize(LAST_MOVE, r.formatLastMove(m))
	}
	r.broadcast(r.serialize(ROOM_STATUS, r.formatRoomStatus()))
}

// unregister terminates the game.DecrementTime routine.
func (r *Room) unregister(c *client) {
	r.Lock()
	defer r.Unlock()

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
		return
	}

	if c.name == r.game.White {
		r.status = WHITE_DISCONNECTED
	} else if c.name == r.game.Black {
		r.status = BLACK_DISCONNECTED
	}
	r.broadcast(r.serialize(ROOM_STATUS, r.formatRoomStatus()))
}

func (r *Room) startGame() {
	r.status = IN_PROGRESS
	go r.game.DecrementTime(r.endGame)

	players := [2]string{}
	i := 0
	for c := range r.clients {
		players[i] = c.name
		i++
	}

	// Randomly select players' sides.
	if rand.Intn(2) == 1 {
		r.game.White = players[0]
		r.game.Black = players[1]
	} else {
		r.game.White = players[1]
		r.game.Black = players[0]
	}
}

func (r *Room) handle(m MoveData, c *client) {
	r.Lock()
	defer r.Unlock()

	if r.status == OPEN || r.status == OVER {
		return
	}

	if !r.isVSEngine {
		if r.game.Bitboard.ActiveColor == enums.White && r.game.White != c.name ||
			r.game.Bitboard.ActiveColor == enums.Black && r.game.Black != c.name {
			return
		}
	}

	if r.game.ProcessMove(bitboard.NewMove(m.To, m.From, m.Type)) {
		lastMove := r.game.Moves[len(r.game.Moves)-1]
		r.broadcast(r.serialize(LAST_MOVE, r.formatLastMove(lastMove)))

		if r.game.Result != enums.Unknown {
			r.endGame()
		}
	}
}

func (r *Room) broadcastChat(data ChatData, c *client) {
	r.Lock()
	defer r.Unlock()

	data.Message = `"` + c.name + `: ` + data.Message + `"`

	msg, err := json.Marshal(Message{Type: CHAT, Data: []byte(data.Message)})
	if err != nil {
		return
	}

	for c := range r.clients {
		c.send <- msg
	}
}

func (r *Room) handleResign(name string) {
	r.Lock()
	defer r.Unlock()

	if r.status != OPEN && r.status != OVER {
		if name == r.game.White {
			r.game.Result = enums.Resignation
			r.game.Winner = enums.Black
			r.endGame()
		} else if name == r.game.Black {
			r.game.Result = enums.Resignation
			r.game.Winner = enums.White
			r.endGame()
		}
	}
}

// endGame ends the game, broadcasts room status and game result.
// endGame cannot be called from the non-locking function.
func (r *Room) endGame() {
	r.status = OVER
	r.broadcast(r.serialize(ROOM_STATUS, r.formatRoomStatus()))

	// Broadcast game result.
	data, _ := json.Marshal(GameResultData{
		Result: r.game.Result,
		Winner: r.game.Winner,
	})
	msg, _ := json.Marshal(Message{Type: GAME_RESULT, Data: data})
	r.broadcast(msg)

	r.hub.remove(r)
}

func (r *Room) formatRoomStatus() RoomStatusData {
	white, black := r.game.White, r.game.Black

	if r.game.White == "" {
		white = "Stockfish 16"
	} else if r.game.Black == "" {
		black = "Stockfish 16"
	}

	return RoomStatusData{
		Status:     r.status,
		White:      white,
		Black:      black,
		WhiteTime:  r.game.WhiteTime,
		BlackTime:  r.game.BlackTime,
		IsVSEngine: r.isVSEngine,
		Clients:    len(r.clients),
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

func (r *Room) getClientIds() []uuid.UUID {
	r.Lock()
	defer r.Unlock()

	ids := make([]uuid.UUID, len(r.clients))
	i := 0
	for c := range r.clients {
		ids[i] = c.id
		i++
	}
	return ids
}
