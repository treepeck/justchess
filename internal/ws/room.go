package ws

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/treepeck/chego"
)

const (
	// Disconnected player has N seconds to reconnect.  If the player doesn't
	// reconnect within the specified time period, victory is awarded to the
	// other player if they are online.  If both players are disconnected the
	// game is marked as abandoned and will not be scored.
	reconnectDeadline int = 60
)

type room struct {
	game               *chego.Game
	moves              []completedMove
	id                 string
	whiteId            string
	blackId            string
	whiteReconnectTime int
	blackReconnectTime int
	clients            map[string]*client
	// When timeToLive is equal to 0, the room will destroy itself.
	register   chan *client
	unregister chan string
	handle     chan event
	clock      *time.Ticker
}

func newRoom(id, whiteId, blackId string) *room {
	return &room{
		id:                 id,
		whiteId:            whiteId,
		blackId:            blackId,
		moves:              make([]completedMove, 0),
		game:               chego.NewGame(),
		whiteReconnectTime: reconnectDeadline,
		blackReconnectTime: reconnectDeadline,
		clients:            make(map[string]*client),
		register:           make(chan *client),
		unregister:         make(chan string),
		handle:             make(chan event),
		clock:              time.NewTicker(time.Second),
	}
}

func (r *room) listenEvents(remove chan<- string) {
	defer func() { remove <- r.id }()

	for {
		select {
		case c := <-r.register:
			r.handleRegister(c)

		case id := <-r.unregister:
			r.handleUnregister(id)

		case e := <-r.handle:
			switch e.Action {
			case actionChat:
				r.handleChat(e)

			case actionMove:
				r.handleMove(e)
				if r.game.IsCheckmate() {
					if len(r.moves)%2 == 0 {
						r.game.Result = chego.BlackWon
					} else {
						r.game.Result = chego.WhiteWon
					}

					r.game.Termination = chego.Normal
				}
			}

		case <-r.clock.C:
			r.handleTimeTick()

			// Destroy the room if both players have been disconnected for a while.
			if r.whiteReconnectTime < 1 && r.blackReconnectTime < 1 {
				return
			} else if r.whiteReconnectTime < 1 {
				// TODO: award the black player with victory.
			} else if r.blackReconnectTime < 1 {
				// TODO: award the white player with victory.
			}
		}
	}
}

func (r *room) handleRegister(c *client) {
	// Deny the connection if the client is already in the queue.
	if _, exist := r.clients[c.player.Id]; exist {
		// Send error event to the client.
		if raw, err := newEncodedEvent(actionError, msgConflict); err == nil {
			c.send <- raw
		} else {
			log.Print(err)
		}
		return
	}

	log.Printf("client %s joined room %s", c.player.Id, r.id)

	c.unregister = r.unregister
	c.forward = r.handle
	r.clients[c.player.Id] = c
	// Send the game state so that the client can sync.
	raw, err := newEncodedEvent(actionGame, gamePayload{
		LegalMoves: r.game.LegalMoves.Moves[:r.game.LegalMoves.LastMoveIndex],
		Moves:      r.moves,
		WhiteTime:  r.game.WhiteTime,
		BlackTime:  r.game.BlackTime,
	})
	if err != nil {
		log.Print(err)
		return
	}
	c.send <- raw

	r.broadcast(actionConn, c.player.Name)
}

func (r *room) handleUnregister(id string) {
	c, exists := r.clients[id]
	if !exists {
		log.Printf("client is not registered")
		return
	}

	log.Printf("client %s leaves room %s", id, r.id)

	delete(r.clients, id)

	r.broadcast(actionDisc, c.player.Name)
}

func (r *room) handleTimeTick() {
	if _, isConnected := r.clients[r.whiteId]; !isConnected {
		r.whiteReconnectTime--
	}

	if _, isConnected := r.clients[r.blackId]; !isConnected {
		r.blackReconnectTime--
	}
}

func (r *room) handleMove(e event) {
	// TODO: Check if it is the player's turn.

	var index byte
	err := json.Unmarshal(e.Payload, &index)
	if err != nil || index >= r.game.LegalMoves.LastMoveIndex {
		e.sender.conn.Close()
		return
	}

	// Store the remaining time.
	tl := r.game.WhiteTime
	if len(r.moves)&2 == 0 {
		tl = r.game.BlackTime
	}

	// Perform and store the move.
	m := r.game.LegalMoves.Moves[index]
	r.moves = append(r.moves, completedMove{
		San:      r.game.PushMove(m),
		Move:     m,
		TimeLeft: tl,
		index:    index,
	})

	r.broadcast(actionMove, movePayload{
		LegalMoves: r.game.LegalMoves.Moves[:r.game.LegalMoves.LastMoveIndex],
		Move:       r.moves[len(r.moves)-1],
	})
}

func (r *room) handleChat(e event) {
	var b strings.Builder
	// Append sender's name.
	b.WriteString(e.sender.player.Name)
	b.WriteString(": ")
	// Append message.
	b.WriteString(strings.TrimSpace(strings.ReplaceAll(string(e.Payload), "\"", " ")))

	e.Payload = json.RawMessage(b.String())
	r.broadcast(actionChat, b.String())
}

// broadcast encodes and sends the event to all connected clients.
func (r *room) broadcast(a eventAction, payload any) {
	raw, err := newEncodedEvent(a, payload)
	if err != nil {
		log.Print(err)
		return
	}

	for _, c := range r.clients {
		c.send <- raw
	}
}
