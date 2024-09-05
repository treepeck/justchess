package ws

import (
	"log"

	"github.com/google/uuid"
)

type room struct {
	Id        uuid.UUID `json:"id"`
	clients   map[*client]bool
	add       chan *client
	remove    chan *client
	broadcast chan event
	isClosed  bool
}

// Creates a new room.
func newRoom() *room {
	return &room{
		Id:        uuid.New(),
		clients:   make(map[*client]bool),
		add:       make(chan *client),
		remove:    make(chan *client),
		broadcast: make(chan event),
		isClosed:  false,
	}
}

func (r *room) run() {
	for {
		select {
		case c := <-r.add:
			r.addClient(c)

		case c := <-r.remove:
			r.removeClient(c)

		case e := <-r.broadcast:
			r.broadcastEvent(e)
		}
	}
}

func (r *room) addClient(c *client) {
	r.clients[c] = true
	log.Println("roomAddClient: clients count: ", len(r.clients))
}

func (r *room) removeClient(c *client) {
	delete(r.clients, c)
	log.Println("roomRemoveClient: clients count: ", len(r.clients))
}

func (r *room) broadcastEvent(e event) {
	for c := range r.clients {
		c.writeEventBuffer <- e
	}
}
