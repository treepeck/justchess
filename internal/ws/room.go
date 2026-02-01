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
	id                 string
	whiteId            string
	blackId            string
	whiteReconnectTime int
	blackReconnectTime int
	clients            map[string]*client
	// When timeToLive is equal to 0, the room will destroy itself.
	register   chan handshake
	unregister chan string
	handle     chan event
	clock      *time.Ticker
}

func newRoom(id, whiteId, blackId string) *room {
	return &room{
		game:               chego.NewGame(),
		id:                 id,
		whiteId:            whiteId,
		blackId:            blackId,
		whiteReconnectTime: reconnectDeadline,
		blackReconnectTime: reconnectDeadline,
		clients:            make(map[string]*client),
		register:           make(chan handshake),
		unregister:         make(chan string),
		handle:             make(chan event),
		clock:              time.NewTicker(time.Second),
	}
}

func (r room) listenEvents(remove chan string) {
	defer func() { remove <- r.id }()

	for {
		select {
		case h := <-r.register:
			r.handleRegister(h)

		case id := <-r.unregister:
			r.handleUnregister(id)

		case e := <-r.handle:
			switch e.Action {
			case actionChat:
				r.handleChat(e)

			case actionMove:
				r.handleMove(e)
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

func (r room) handleRegister(h handshake) {
	// Deny the connection if the client is already in the room.
	if _, exist := r.clients[h.player.Id]; exist {
		h.isConflict <- true
		return
	}

	conn, err := upgrader.Upgrade(h.rw, h.r, nil)
	h.isConflict <- false
	if err != nil {
		// upgrader writes the response, so simply return here.
		return
	}

	c := newClient(conn, h.player)
	go c.read(r.unregister, r.handle)
	go c.write()

	r.clients[h.player.Id] = c
	log.Printf("client %s joined room %s", h.player.Id, r.id)
}

func (r room) handleUnregister(id string) {
	_, exists := r.clients[id]
	if !exists {
		log.Printf("client is not registered")
		return
	}

	delete(r.clients, id)
	log.Printf("client %s leaves room %s", id, r.id)
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
	// Check if it is the player's turn.

}

func (r room) handleChat(e event) {
	var b strings.Builder
	// Append opening quote.
	b.WriteByte('"')
	// Append sender's name.
	b.WriteString(e.sender.player.Name)
	b.WriteString(": ")
	// Append message.
	b.WriteString(strings.TrimSpace(strings.ReplaceAll(string(e.Payload), "\"", " ")))
	// Append final quote.
	b.WriteByte('"')

	e.Payload = json.RawMessage(b.String())
	r.broadcast(e)
}

// broadcast event among all connected clients.  It's the caller's responsibility
// to encode the event payload.
func (r room) broadcast(e event) {
	raw, err := json.Marshal(e)
	if err != nil {
		log.Print(err)
		return
	}

	for _, c := range r.clients {
		c.send <- raw
	}
}
