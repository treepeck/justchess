package ws

import (
	"chess-api/repository"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// upgrader is used by the Manager to recieve a *Conn.
var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		// All connections except front-end are prohibited.
		// To test a ws package, this function must return true always.
		return r.Header.Get("Origin") == os.Getenv("CLIENT_DOMAIN")
		// return true // uncomment while testing a ws package.
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
// If the connection cannot be upgraded, it sends a header with the status code 500
// back to the client.
func (m *Manager) HandleConnection(rw http.ResponseWriter, r *http.Request) {
	fn := slog.String("func", "HandleConnection")
	// TODO: replace with the Authorization and take AccessToken From Authorization Header
	idStr := r.URL.Query().Get("id")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		slog.Warn("cannot parse uuid", fn, "err", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	u := repository.FindUserById(userId)
	if u == nil {
		slog.Warn("user not found", fn, "err", err)
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		slog.Warn("error while upgrading the connection", fn, "err", err)
		return
	}

	c := newClient(conn, m, *u)
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
	fn := slog.String("func", "manager.addClient")

	// if the client is alredy connected, close the previous connection.
	for connC := range m.clients {
		if connC.User.Id == c.User.Id {
			m.removeClient(connC)
		}
	}

	m.clients[c] = true
	slog.Info("client "+c.User.Name+" joined", fn)

	go c.readEvents()
	go c.writeEvents()

	m.broadcastCC()
}

// removeClient removes client from the clients map.
// Closes a connection with the front-end.
func (m *Manager) removeClient(c *Client) {
	fn := slog.String("func", "manager.removeClient")

	if _, ok := m.clients[c]; ok {
		c.conn.Close()
		delete(m.clients, c)

		slog.Info("client "+c.User.Name+" removed", fn)
		m.broadcastCC()

		if c.currentRoom != nil {
			c.currentRoom.unregister <- c
		}
	}
}

// addRoom adds a new room.
func (m *Manager) addRoom(r *Room) {
	fn := slog.String("func", "addRoom")
	m.rooms[r] = true
	slog.Info("room added", fn, slog.Int("count", len(m.rooms)))
	m.broadcastAddRoom(r)
}

// removeRoom removes a room.
func (m *Manager) removeRoom(r *Room) {
	fn := slog.String("func", "removeRoom")

	if _, ok := m.rooms[r]; ok {
		delete(m.rooms, r)
		r.close <- true // end room goroutine
		slog.Info("room removed", fn, slog.Int("count", len(m.rooms)))
	}

	m.broadcastRemoveRoom(r)
}

// findRoomById finds the room with the specified id.
func (m *Manager) findRoomById(id uuid.UUID) *Room {
	for r := range m.rooms {
		if r.Id == id {
			return r
		}
	}
	return nil
}

// broadcastAddRoom broadcasts the added room.
func (m *Manager) broadcastAddRoom(r *Room) {
	fn := slog.String("func", "broadcastAddRoom")

	p, err := json.Marshal(r)
	if err != nil {
		slog.Warn("cannot Marshal Room", fn, "err", err)
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
	fn := slog.String("func", "broadcastRemoveRoom")

	p, err := json.Marshal(r)
	if err != nil {
		slog.Warn("cannot Marshal Room", fn, "err", err)
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
	fn := slog.String("func", "broadcastCC")

	p, err := json.Marshal(len(m.clients))
	if err != nil {
		slog.Warn("cannot Marshal clients counter", fn, "err", err)
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
