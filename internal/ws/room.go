package ws

import (
	"encoding/json"
	"log"
	"time"

	"justchess/internal/db"

	"github.com/treepeck/chego"
)

// Empty room will live 20 seconds before destruction.
const emptyDeadline int = 20

type room struct {
	game    *chego.Game
	id      string
	whiteId string
	blackId string
	clients map[*client]db.Player
	// When timeToLive is equal to 0, the room will destroy itself.
	timeToLive int
	register   chan handshake
	unregister chan *client
	handle     chan event
	clock      *time.Ticker
}

func newRoom(id, whiteId, blackId string) room {
	return room{
		game:       chego.NewGame(),
		id:         id,
		whiteId:    whiteId,
		blackId:    blackId,
		clients:    make(map[*client]db.Player),
		timeToLive: emptyDeadline,
		register:   make(chan handshake),
		unregister: make(chan *client),
		handle:     make(chan event),
		clock:      time.NewTicker(time.Second),
	}
}

func (r room) listenEvents(remove chan string) {
	defer func() { remove <- r.id }()

	for {
		select {
		case h := <-r.register:
			r.handleRegister(h)

		case c := <-r.unregister:
			r.handleUnregister(c)

		case e := <-r.handle:
			switch e.Action {
			case actionChat:
				r.handleChat(e)
			}

		case <-r.clock.C:
			r.handleTimeTick()

			if r.timeToLive == 0 {
				// Destroy empty room.
				return
			}
		}
	}
}

func (r room) handleRegister(h handshake) {
	// Write to the response channel so that request cannot be closed.
	defer func() { h.ch <- struct{}{} }()

	conn, err := upgrader.Upgrade(h.rw, h.r, nil)
	if err != nil {
		// upgrader writes the response, so simply return here.
		return
	}

	c := newClient(conn)
	go c.read(r.unregister, nil)
	go c.write()

	r.clients[c] = h.player
	log.Printf("client %s joined room %s", h.player.Id, r.id)
}

func (r room) handleUnregister(c *client) {
	p, exists := r.clients[c]
	if !exists {
		log.Printf("client is not registered")
		return
	}

	delete(r.clients, c)
	log.Printf("client %s leaves room %s", p.Id, r.id)
}

func (r room) handleTimeTick() {
	// If there are no players in the room, decremen the ttl.
	if len(r.clients) == 0 {

	}

}

// broadcasts chat message.
// TODO: sanitize chat messages.
func (r room) handleChat(e event) {
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

	for c := range r.clients {
		c.send <- raw
	}
}
