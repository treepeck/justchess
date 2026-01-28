package ws

import (
	"encoding/json"
	"log"

	"github.com/treepeck/chego"
)

type room struct {
	game       *chego.Game
	id         string
	whiteId    string
	blackId    string
	clients    map[string]*client
	register   chan handshake
	unregister chan *client
	handle     chan event
}

func newRoom(id, whiteId, blackId string) room {
	return room{
		game:       chego.NewGame(),
		id:         id,
		whiteId:    whiteId,
		blackId:    blackId,
		clients:    make(map[string]*client),
		register:   make(chan handshake),
		unregister: make(chan *client),
		handle:     make(chan event),
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
		}
	}
}

func (r room) handleRegister(h handshake) {
	// Write to the response channel so that request cannot be closed.
	defer func() { h.ch <- struct{}{} }()

	// Deny the request if the client is already in the room.
	if _, exists := r.clients[h.player.Id]; exists {
		return
	}

	conn, err := upgrader.Upgrade(h.rw, h.r, nil)
	if err != nil {
		// upgrader writes the response, so simply return here.
		return
	}

	c := newClient(h.player, conn)
	go c.read(r.unregister, nil)
	go c.write()

	r.clients[h.player.Id] = c
	log.Printf("client %s joined room %s", h.player.Id, r.id)
}

func (r room) handleUnregister(c *client) {
	if _, exists := r.clients[c.player.Id]; !exists {
		log.Printf("client %s is not registered", c.player.Id)
		return
	}

	delete(r.clients, c.player.Id)
	log.Printf("client %s leaved room %s", c.player.Id, r.id)
}

// broadcasts chat message.
// TODO: sanitize and rate limir messages.
func (r room) handleChat(e event) {
	r.broadcast(e)
}

// broadcast even among all connected clients.  It's the caller's responsibility
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
