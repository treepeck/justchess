package ws

import (
	"encoding/json"
	"justchess/pkg/auth"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// upgrader is used by the Manager to recieve a *Conn.
var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Manager stores the map of connected clients, handles new connections and
// disconnections.
type Manager struct {
	register   chan *Client
	unregister chan *Client
	add        chan *Room
	remove     chan *Room
	broadcast  chan Event
	clients    map[*Client]bool
	rooms      map[*Room]bool
}

// NewManager creates and runs a new manager.
func NewManager() *Manager {
	m := &Manager{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		add:        make(chan *Room),
		remove:     make(chan *Room),
		broadcast:  make(chan Event),
		clients:    make(map[*Client]bool),
		rooms:      make(map[*Room]bool),
	}
	go m.run()
	return m
}

// HandleConnection upgrades the incoming HTTP connection to the WebSocket Protocol.
// To be connected, client must provide a valid access JWT as a request param.
func (m *Manager) HandleConnection(rw http.ResponseWriter, r *http.Request) {
	et := r.URL.Query().Get("at")
	at, err := auth.DecodeToken(et, "ATS")
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	idStr, err := at.Claims.GetSubject()
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
		slog.Warn("error while upgrading the connection", "err", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	c := newClient(conn, m, id)
	// the register channel is used to avoid concurrent access to the clients map.
	m.register <- c
}

// Run receives data via channels and processes it.
func (m *Manager) run() {
	for {
		select {
		case c := <-m.register:
			m.addClient(c)

		case c := <-m.unregister:
			m.removeClient(c)

		case r := <-m.add:
			m.addRoom(r)

		case r := <-m.remove:
			m.removeRoom(r)

		case e := <-m.broadcast:
			// broadcast the event among all connected clients.
			for c := range m.clients {
				// skip the clients that are already in a game.
				if c.currentRoom == nil {
					c.writeEventBuffer <- e
				}
			}
		}
	}
}

// addClient adds a new client to the clients map and invokes the client`s goroutines.
func (m *Manager) addClient(c *Client) {
	// if the client is alredy connected, close the previous connection.
	for connC := range m.clients {
		if connC.Id == c.Id {
			m.removeClient(connC)
		}
	}

	m.clients[c] = true
	slog.Info("client " + c.Id.String() + " joined")

	go c.readEvents()
	go c.writeEvents()

	m.broadcastCC()
}

// removeClient removes client from the clients map.
// Closes a connection with the front-end.
func (m *Manager) removeClient(c *Client) {
	if _, ok := m.clients[c]; ok {
		if c.currentRoom != nil {
			c.currentRoom.unregister <- c
		}

		c.conn.Close()
		delete(m.clients, c)

		slog.Info("client " + c.Id.String() + " removed")
		m.broadcastCC()
	}
}

// addRoom adds a new room.
func (m *Manager) addRoom(r *Room) {
	m.rooms[r] = true
	slog.Info("room added", slog.Int("count", len(m.rooms)))
	m.broadcastAddRoom(r)
}

// removeRoom removes a room.
func (m *Manager) removeRoom(r *Room) {
	if _, ok := m.rooms[r]; ok {
		delete(m.rooms, r)
		r.close <- true // exit room Run loop
		slog.Info("room removed", slog.Int("count", len(m.rooms)))
		m.broadcastRemoveRoom(r)
	}
}

// findRoomById finds the room with the specified id.
func (m *Manager) findRoomById(id uuid.UUID) *Room {
	for r := range m.rooms {
		if r.id == id {
			return r
		}
	}
	return nil
}

// broadcastAddRoom broadcasts the added room.
func (m *Manager) broadcastAddRoom(r *Room) {
	p, err := json.Marshal(r)
	if err != nil {
		slog.Warn("cannot Marshal Room", "err", err)
		return
	}
	e := Event{
		Action:  ADD_ROOM,
		Payload: p,
	}
	for c := range m.clients {
		// skip the clients that are already in a game.
		if c.currentRoom == nil {
			c.writeEventBuffer <- e
		}
	}
}

// broadcastRemoveRoom is a helper function that broadcasts the removed room.
func (m *Manager) broadcastRemoveRoom(r *Room) {
	p, err := json.Marshal(r)
	if err != nil {
		slog.Warn("cannot Marshal Room", "err", err)
		return
	}
	e := Event{
		Action:  REMOVE_ROOM,
		Payload: p,
	}
	for c := range m.clients {
		// skip the clients that are already in a game.
		if c.currentRoom == nil {
			c.writeEventBuffer <- e
		}
	}
}

// broadcastCC is a helper function that broadcasts the updated clients counter.
func (m *Manager) broadcastCC() {
	p, err := json.Marshal(len(m.clients))
	if err != nil {
		slog.Warn("cannot Marshal clients counter", "err", err)
		return
	}
	e := Event{
		Action:  CLIENTS_COUNTER,
		Payload: p,
	}
	for c := range m.clients {
		// skip the clients that are already in a game.
		if c.currentRoom == nil {
			c.writeEventBuffer <- e
		}
	}
}
