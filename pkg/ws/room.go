package ws

import (
	"encoding/json"
	"justchess/pkg/auth"
	"justchess/pkg/chess"
	"justchess/pkg/chess/bitboard"
	"justchess/pkg/chess/enums"
	"justchess/pkg/game"
	"log"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
)

type roomStatus int
type clientEventType int

type moveEvent struct {
	client *client
	move   MoveDTO
}

type clientEvent struct {
	client *client
	eType  clientEventType
}

type chatEvent struct {
	client *client
	data   ChatDTO
}

const (
	statusOpen roomStatus = iota
	statusInProgress
	statusWhiteDisconnected
	statusBlackDisconnected
	statusOver

	typeRegister clientEventType = iota - 5
	typeUnregister
	typeResign
	typeOfferDraw
	typeDeclineDraw
)

type Room struct {
	Id                uuid.UUID  `json:"id"`
	CreatorName       string     `json:"cn"`
	Status            roomStatus `json:"s"`
	IsVSEngine        bool       `json:"e"` // Whet	her the match is between player and engine.
	pendingDrawIssuer uuid.UUID
	hub               *Hub // To be able to remove the room from the hub.
	game              *chess.Game
	clients           []*client
	timeout           chan struct{} // The game will send to timeout if one of the players runns out of time.
	clientEvents      chan clientEvent
	move              chan moveEvent
	chat              chan chatEvent
}

func newRoom(h *Hub, creatorName string, isVSEngine bool, control, bonus int) *Room {
	id := uuid.New()
	return &Room{
		Id:           id,
		CreatorName:  creatorName,
		Status:       statusOpen,
		IsVSEngine:   isVSEngine,
		hub:          h,
		game:         chess.NewGame("", control, bonus),
		clients:      make([]*client, 0),
		timeout:      make(chan struct{}),
		clientEvents: make(chan clientEvent),
		move:         make(chan moveEvent),
		chat:         make(chan chatEvent),
	}
}

func (r *Room) runRoutine() {
	for {
		select {
		case e := <-r.clientEvents:
			switch e.eType {
			case typeRegister:
				r.register(e.client)
			case typeUnregister:
				r.unregister(e.client)

				if len(r.clients) == 0 {
					return
				}
			case typeResign:
				r.resign(e.client)
			case typeOfferDraw:
				r.offerDraw(e.client)
			case typeDeclineDraw:
				r.declineDraw(e.client)
			}

		case me := <-r.move:
			r.handleMove(me)

		case ce := <-r.chat:
			r.broadcastChat(ce.data, ce.client.name)

		case <-r.timeout:
			r.endGame()
		}
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
	if !r.IsVSEngine && cms.Role == auth.RoleGuest {
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

	r.clientEvents <- clientEvent{client: c, eType: typeRegister}
}

// register registers the client in the room. Repeated connections from the single peer are denied.
func (r *Room) register(c *client) {
	// Deny connection if the client is already connected or the game is already over.
	for _, conn := range r.clients {
		if conn.id == c.id || r.Status == statusOver {
			close(c.send)
			return
		}
	}

	r.clients = append(r.clients, c)
	log.Printf("client %s registered\n", c.id.String())

	switch r.Status {
	// Start the game if both clients are connected or if the math is between player and engine.
	case statusOpen:
		if r.IsVSEngine || len(r.clients) == 2 {
			r.startGame()
		}

	case statusWhiteDisconnected:
		if r.game.WhiteId == c.id {
			r.Status = statusInProgress
		}

	case statusBlackDisconnected:
		if r.game.BlackId == c.id {
			r.Status = statusInProgress
		}
	}

	r.broadcastRoomStatus()
}

func (r *Room) unregister(c *client) {
	for i, conn := range r.clients {
		if conn.id == c.id {
			close(c.send)
			// Remove the client. To do that, replace the element to delete with the one
			// at the end and then assign to the len-1 elements.
			r.clients[i] = r.clients[len(r.clients)-1]
			r.clients = r.clients[:len(r.clients)-1]

			log.Printf("client %s unregistered\n", c.id.String())
			if len(r.clients) == 0 {
				if r.Status != statusOpen && r.Status != statusOver {
					r.game.End <- struct{}{}
					r.game.SetEndInfo(enums.Unknown, enums.None)
				}

				r.hub.remove(r)
				return
			}

			if c.id == r.game.WhiteId && r.Status != statusOver {
				r.Status = statusWhiteDisconnected
			} else if c.id == r.game.BlackId && r.Status != statusOver {
				r.Status = statusBlackDisconnected
			}

			r.broadcastRoomStatus()
			break
		}
	}

}

func (r *Room) handleMove(m moveEvent) {
	isPlayerTurn := (r.game.Bitboard.ActiveColor == enums.White && r.game.WhiteId == m.client.id) ||
		(r.game.Bitboard.ActiveColor == enums.Black && r.game.BlackId == m.client.id)

	if r.Status == statusOpen || r.Status == statusOver ||
		(r.game.WhiteId != m.client.id && r.game.BlackId != m.client.id) ||
		(!r.IsVSEngine && !isPlayerTurn) {
		return
	}

	res := make(chan bool)
	r.game.Move <- chess.MoveEvent{
		ClientId: m.client.id,
		Move:     bitboard.NewMove(m.move.Destination, m.move.Source, m.move.Type),
		Response: res,
	}

	// If the move wasn't processed.
	if !<-res {
		return
	}

	lastMove := r.game.Moves[len(r.game.Moves)-1]
	data, err := json.Marshal(Message{Type: LAST_MOVE, Data: r.serializeLastMove(lastMove, r.game.Bitboard.LegalMoves)})
	if err != nil {
		log.Printf("cannot Marshal data: %v\n", err)
		return
	}
	r.broadcast(data)

	if r.game.Result != enums.Unknown {
		r.endGame()
	}
}

func (r *Room) resign(c *client) {
	if r.Status == statusOpen || r.Status == statusOver {
		return
	}

	if c.id == r.game.WhiteId {
		r.game.End <- struct{}{}
		r.game.SetEndInfo(enums.Resignation, enums.Black)
		r.endGame()
	} else if c.id == r.game.BlackId {
		r.game.End <- struct{}{}
		r.game.SetEndInfo(enums.Resignation, enums.White)
		r.endGame()
	}
}

func (r *Room) offerDraw(c *client) {
	if r.Status == statusOpen || r.Status == statusOver {
		return
	}

	if r.pendingDrawIssuer == uuid.Nil {
		var oppId uuid.UUID
		var msg []byte

		if c.id == r.game.WhiteId {
			oppId = r.game.BlackId
			msg, _ = json.Marshal(Message{Type: CHAT, Data: []byte(`"White offers draw"`)})
		} else if c.id == r.game.BlackId {
			oppId = r.game.WhiteId
			msg, _ = json.Marshal(Message{Type: CHAT, Data: []byte(`"Black offers draw"`)})
		} else { // Deny draw offers from viewers.
			return
		}

		r.broadcast(msg)
		r.pendingDrawIssuer = c.id

		msg, _ = json.Marshal(Message{Type: DRAW_OFFER, Data: nil})
		for _, conn := range r.clients {
			if conn.id == oppId {
				conn.send <- msg
			}
		}
	} else if r.pendingDrawIssuer != c.id && (c.id == r.game.WhiteId ||
		c.id == r.game.BlackId) {
		msg, _ := json.Marshal(Message{Type: CHAT, Data: []byte(`"Draw accepted"`)})
		r.broadcast(msg)

		r.game.End <- struct{}{}
		r.game.SetEndInfo(enums.Agreement, enums.None)
		r.endGame()
	}
}

func (r *Room) declineDraw(c *client) {
	if r.Status == statusOpen || r.Status == statusOver ||
		c.id == r.pendingDrawIssuer || r.pendingDrawIssuer == uuid.Nil ||
		(c.id != r.game.WhiteId && c.id != r.game.BlackId) {
		return
	}

	msg, _ := json.Marshal(Message{Type: CHAT, Data: []byte(`"Draw declined"`)})
	r.broadcast(msg)

	r.pendingDrawIssuer = uuid.Nil
}

func (r *Room) broadcastChat(data ChatDTO, senderName string) {
	data.Message = `"` + senderName + `: ` + data.Message + `"`

	msg, err := json.Marshal(Message{Type: CHAT, Data: []byte(data.Message)})
	if err != nil {
		return
	}
	r.broadcast(msg)
}

func (r *Room) startGame() {
	r.Status = statusInProgress

	go r.game.RunRoutine(r.timeout)

	if rand.Intn(2) == 1 {
		r.game.WhiteId = r.clients[0].id
		if len(r.clients) == 2 {
			r.game.BlackId = r.clients[1].id
		} else {
			// TODO: unsafe code, the stockfish' uuid cannot ever change without breaking this.
			r.game.BlackId = uuid.MustParse("ccaf962b-855e-49da-b85f-7e8bba0edae2")
		}
	} else {
		r.game.BlackId = r.clients[0].id
		if len(r.clients) == 2 {
			r.game.WhiteId = r.clients[1].id
		} else {
			r.game.WhiteId = uuid.MustParse("ccaf962b-855e-49da-b85f-7e8bba0edae2")
		}
	}
}

// endGame broadcasts room status and the game result.
// Game data is stored in the db.
func (r *Room) endGame() {
	r.Status = statusOver

	data, _ := json.Marshal(GameResultDTO{Result: r.game.Result, Winner: r.game.Winner})
	msg, _ := json.Marshal(Message{Type: GAME_RESULT, Data: data})
	r.broadcast(msg)

	r.hub.remove(r)

	if len(r.game.Moves) > 1 {
		err := game.Insert(r.Id.String(), *r.game)
		if err != nil {
			log.Printf("cannot store game in the db: %v, game: %v\n", err, *r.game)
		}
	}
}

func (r *Room) broadcast(msg []byte) {
	for _, c := range r.clients {
		c.send <- msg
	}
}

func (r *Room) broadcastRoomStatus() {
	info := chess.GameInfo{}

	if r.Status == statusOpen || r.Status == statusOver {
		info = chess.GameInfo{
			WhiteTime:  r.game.WhiteTime,
			BlackTime:  r.game.BlackTime,
			Result:     r.game.Result,
			Winner:     r.game.Winner,
			Moves:      r.game.Moves,
			LegalMoves: r.game.Bitboard.LegalMoves,
		}
	} else {
		res := make(chan chess.GameInfo)
		r.game.Info <- chess.GameInfoEvent{Response: res}
		info = <-res
	}
	data, err := json.Marshal(Message{Type: ROOM_STATUS, Data: r.serializeRoomStatus(info)})
	if err != nil {
		log.Printf("cannot Marshal data: %v\n", err)
		return
	}
	r.broadcast(data)
}

func (r *Room) serializeRoomStatus(info chess.GameInfo) []byte {
	data, err := json.Marshal(RoomStatusDTO{
		Status:     r.Status,
		White:      r.game.WhiteId,
		Black:      r.game.BlackId,
		WhiteTime:  info.WhiteTime,
		BlackTime:  info.BlackTime,
		Control:    r.game.TimeControl,
		IsVSEngine: r.IsVSEngine,
		Clients:    len(r.clients),
	})
	if err != nil {
		log.Printf("cannot Marshal room status: %v\n", err)
	}
	return data
}

func (r *Room) serializeLastMove(move chess.CompletedMove, lm []bitboard.Move) []byte {
	data, err := json.Marshal(LastMoveDTO{
		SAN:        move.SAN,
		FEN:        move.FEN,
		TimeLeft:   move.TimeLeft,
		LegalMoves: lm,
	})
	if err != nil {
		log.Printf("cannot Marshal last move: %v\n", err)
	}
	return data
}
