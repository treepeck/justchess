package ws

import (
	"justchess/pkg/auth"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// upgrader is used to upgrate the HTTP connection into the websocket protocol.
var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	// CheckOrigin accepts all origins since the CORS is handled by the corsAllower middleware.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Hub is responsible for creating and deleting the rooms.
type Hub struct {
	register   chan *client
	unregister chan *client
	add        chan *Room
	remove     chan uuid.UUID
	clients    map[*client]struct{}
	rooms      map[uuid.UUID]*Room
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *client),
		unregister: make(chan *client),
		add:        make(chan *Room),
		remove:     make(chan uuid.UUID),
		clients:    make(map[*client]struct{}),
		rooms:      make(map[uuid.UUID]*Room),
	}
}

func (h *Hub) HandleNewConnection(rw http.ResponseWriter, r *http.Request) {
	encoded := r.URL.Query().Get("access")
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

	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	c := newClient(id, conn)
	c.hub = h

	go c.readPump()
	go c.writePump()

	h.register <- c
}

func (h *Hub) EventPump() {
	for {
		select {
		case c := <-h.register:
			h.registerClient(c)
			h.broadcastClientsCounter()

		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				h.unregisterClient(c)
			}

		case r := <-h.add:
			h.addRoom(r)
			h.broadcastAddRoom(r)

		case id := <-h.remove:
			if _, ok := h.rooms[id]; ok {
				h.removeRoom(id)
			}
		}
	}
}

func (h *Hub) registerClient(c *client) {
	for connected := range h.clients {
		if connected.id == c.id {
			// Deny multiple connections from a single peer.
			close(c.send)
			return
		}
	}

	h.clients[c] = struct{}{}
	log.Printf("client %s added\n", c.id.String())
}

func (h *Hub) unregisterClient(c *client) {
	delete(h.clients, c)
	log.Printf("client %s removed\n", c.id.String())
}

func (h *Hub) addRoom(r *Room) {
	h.rooms[r.creatorId] = r
	log.Printf("user %s created a room\n", r.creatorId.String())
}

func (h *Hub) removeRoom(id uuid.UUID) {
	delete(h.rooms, id)
	log.Printf("room created by %s removed\n", id.String())
}

// broadcastClientsCounter sends clients counter to all connected clients.
// To send larger numbers, such as uint32, the message size is 5 bytes.
func (h *Hub) broadcastClientsCounter() {
	// TODO: handle larger numbers, than 256.
	msg := []byte{byte(len(h.clients)), CLIENTS_COUNTER}

	for c := range h.clients {
		c.send <- msg
	}
}

// broadcastAddRoom sends room info to all connected clients.
func (h *Hub) broadcastAddRoom(r *Room) {
	msg := make([]byte, 19)
	copy(msg[:16], r.creatorId[:])
	msg[16] = r.game.TimeControl
	msg[17] = r.game.TimeBonus
	msg[18] = ADD_ROOM

	for c := range h.clients {
		c.send <- msg
	}
}
