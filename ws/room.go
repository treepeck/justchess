package ws

import (
	"log/slog"

	"github.com/google/uuid"
)

type room struct {
	Id        uuid.UUID `json:"id"`
	Control   string    `json:"control"`
	Bonus     uint      `json:"bonus"`
	Rating    uint      `json:"rating"`
	clients   map[*client]bool
	add       chan *client
	remove    chan *client
	broadcast chan event
	isClosed  bool
}

type CreateRoomDTO struct {
	Control string `json:"control"`
	Bonus   uint   `json:"bonus"`
	Rating  uint   `json:"rating"`
}

// Creates a new room.
func newRoom(cr CreateRoomDTO) *room {
	return &room{
		Id:        uuid.New(),
		Rating:    cr.Rating,
		Control:   cr.Control,
		Bonus:     cr.Bonus,
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
	if !r.isClosed {
		fn := slog.String("func", "room.addClient")

		r.clients[c] = true
		slog.Info("client added", fn, slog.Int("counter", len(r.clients)))

		if len(r.clients) > 1 {
			r.isClosed = true
		}
	}
}

func (r *room) removeClient(c *client) {
	fn := slog.String("func", "room.removeClient")
	delete(r.clients, c)
	slog.Info("client removed", fn, slog.Int("counter", len(r.clients)))
}

func (r *room) broadcastEvent(e event) {
	for c := range r.clients {
		c.writeEventBuffer <- e
	}
}
