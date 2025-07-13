package ws

import (
	"encoding/json"
	"log"
	"math/rand"

	"github.com/BelikovArtem/chego/enum"
	"github.com/BelikovArtem/chego/game"
)

type roomState int

const (
	stateInProgress roomState = iota
	stateOver
)

// Room is a middleman between a single game and connected clients.
type Room struct {
	// Connected clients.
	clients map[*client]struct{}
	// Channel to register and unregister clients.
	gate chan *client
	// Channel to handle moves.
	move chan makeMovePayload
	game *game.Game
	// Ends eventRoutine.
	destroy          chan struct{}
	state            roomState
	isWhiteConnected bool
	isBlackConnected bool
}

func NewRoom() *Room {
	return &Room{
		clients:          make(map[*client]struct{}),
		gate:             make(chan *client),
		move:             make(chan makeMovePayload),
		game:             nil,
		destroy:          make(chan struct{}),
		state:            stateOver,
		isWhiteConnected: false,
		isBlackConnected: false,
	}
}

// EventRoutine starts endless loop to safely handle concurrent events emmited by clients.
func (r *Room) EventRoutine() {
	for {
		select {
		case c := <-r.gate:
			if _, ok := r.clients[c]; ok {
				r.unregister(c)
			} else {
				r.register(c)
			}

		case m := <-r.move:
			r.handleMove(m)

		case <-r.destroy:
			// Disconnect all clients before destroing the room.
			for c := range r.clients {
				r.unregister(c)
			}
			log.Printf("Room destroyed")
			return
		}
	}
}

// register registers the client in the room and starts the game if
// there is enough players.
func (r *Room) register(c *client) {
	log.Printf("Registered %s", c.id)
	r.clients[c] = struct{}{}

	// TODO: remove later.
	if len(r.clients) == 1 {
		c.isRoomCreator = true
	}

	if c.isRoomCreator {
		switch c.color {
		case colorWhite:
			r.isWhiteConnected = true
		case colorBlack:
			r.isBlackConnected = true
		default:
			// Randomly pick player color.
			if rand.Intn(2) == 1 {
				r.isWhiteConnected = true
				c.color = colorWhite
			} else {
				r.isBlackConnected = true
				c.color = colorBlack
			}
		}
	} else if len(r.clients) == 2 {
		if !r.isWhiteConnected {
			r.isWhiteConnected = true
			c.color = colorWhite
		} else {
			r.isBlackConnected = true
			c.color = colorBlack
		}

		r.startGame()
	}

	r.broadcastRoomInfo()
}

// unregister unregisters the client.
// TODO: consider destroing the room if there are no clients left.
func (r *Room) unregister(c *client) {
	log.Printf("Unregistered %s", c.id)
	delete(r.clients, c)

	switch c.color {
	case colorWhite:
		r.isWhiteConnected = false
	case colorBlack:
		r.isBlackConnected = false
	}

	r.broadcastRoomInfo()
}

func (r *Room) startGame() {
	log.Printf("Game started")
	r.state = stateInProgress

	r.game = game.NewGame()
}

func (r *Room) endGame(res enum.Result) {
	log.Printf("Game ended with result %d", res)
	r.state = stateOver

	r.game = nil
}

func (r *Room) handleMove(m makeMovePayload) {
	if r.state != stateInProgress {
		log.Printf("Recieved move message but the game is not in progress %s", m.senderId)
		return
	}

	log.Printf("Recieved move message %s", m.senderId)
	legalMoveIndex := r.game.GetLegalMoveIndex(m.To, m.From, m.PromotionPiece)
	if legalMoveIndex == -1 {
		return
	}
	r.game.PushMove(r.game.LegalMoves.Moves[legalMoveIndex])

	r.broadcastLastMove()

	if r.game.IsInsufficientMaterial() {
		r.endGame(enum.ResultInsufficientMaterial)
	} else if r.game.IsThreefoldRepetition() {
		r.endGame(enum.ResultThreefoldRepetition)
	} else if r.game.IsCheckmate() {
		r.endGame(enum.ResultCheckmate)
	} else if r.game.LegalMoves.LastMoveIndex == 0 {
		r.endGame(enum.ResultStalemate)
	}
}

// broadcastRoomInfo broadcasts the room info among all connected clients.
func (r *Room) broadcastRoomInfo() {
	p, err := json.Marshal(roomInfoPayload{
		State:            r.state,
		ClientsCounter:   len(r.clients),
		IsWhiteConnected: r.isWhiteConnected,
		IsBlackConnected: r.isBlackConnected,
	})
	if err != nil {
		log.Printf("Cannot Marshal room info: %v", err)
		return
	}

	msg, _ := json.Marshal(message{Action: actionRoomInfo, Payload: string(p)})

	for c := range r.clients {
		c.send <- msg
	}
}

// broadcastLastMove broadcasts the last completed move among all connected clients.
func (r *Room) broadcastLastMove() {
	lastMove := r.game.MoveStack[len(r.game.MoveStack)-1]

	p, err := json.Marshal(lastMovePayload{
		FenString:  getPiecePlacementData(lastMove.FenString),
		LegalMoves: r.compressLegalMoves(),
	})
	if err != nil {
		log.Printf("Cannot Marshal last move: %v\n", err)
		return
	}

	msg, _ := json.Marshal(message{Action: actionLastMove, Payload: string(p)})

	for c := range r.clients {
		c.send <- msg
	}
}

func (r *Room) broadcastGameOver() {

}

// Used to compress legal moves.
var squareString = [64]string{
	"a1", "b1", "c1", "d1", "e1", "f1", "g1", "h1",
	"a2", "b2", "c2", "d2", "e2", "f2", "g2", "h2",
	"a3", "b3", "c3", "d3", "e3", "f3", "g3", "h3",
	"a4", "b4", "c4", "d4", "e4", "f4", "g4", "h4",
	"a5", "b5", "c5", "d5", "e5", "f5", "g5", "h5",
	"a6", "b6", "c6", "d6", "e6", "f6", "g6", "h6",
	"a7", "b7", "c7", "d7", "e7", "f7", "g7", "h7",
	"a8", "b8", "c8", "d8", "e8", "f8", "g8", "h8",
}

// compressLegalMoves is a helper function to encode legal moves into the compact string.
func (r *Room) compressLegalMoves() map[string]string {
	destinations := make(map[string]string)
	var i byte
	for i = 0; i < r.game.LegalMoves.LastMoveIndex; i++ {
		to := squareString[r.game.LegalMoves.Moves[i].To()]
		from := squareString[r.game.LegalMoves.Moves[i].From()]

		destinations[from] += to
	}
	return destinations
}

// getFenPiecePlacementData is a helper function to extract the piece placement data
// (1 field) from the FEN string.
func getPiecePlacementData(fenStr string) string {
	for i := 0; i < len(fenStr); i++ {
		if fenStr[i] == ' ' {
			return fenStr[:i]
		}
	}
	panic("FEN string only contain one field")
}
