package ws

import (
	"chess-api/repository"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin: func(req *http.Request) bool {
		return true
	},
}

type Manager struct {
	sync.Mutex
	Clients map[*client]bool
	Rooms   map[*room]bool
}

// Creates a new Manager.
func NewManager() *Manager {
	return &Manager{
		Clients: make(map[*client]bool),
		Rooms:   make(map[*room]bool),
	}
}

// Upgrades the incoming HTTP connection to the WebSocket Protocol.
// If the connection cannot be upgraded, sends a header with status code 500
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
	m.addClient(c)
}

// Adds a new client to the clients map and invokes the client`s goroutines:
//  1. readEvents goroutine handles the incomming events from the client;
//  2. writeEvent goroutine grabs the events from the evBuf channel and sends those
//     events to the client.
func (m *Manager) addClient(c *client) {
	fn := slog.String("func", "addClient")

	m.Lock()
	defer m.Unlock()

	m.Clients[c] = true
	slog.Info("client "+c.user.GetName()+" joined", fn)

	go c.readEvents()
	go c.writeEvents()

	m.broadcast(UPDATE_CLIENTS_COUNTER)
}

// Removes client from the clients map. Closes a connection with the front-end.
func (m *Manager) removeClient(c *client) {
	fn := slog.String("func", "removeClient")

	m.Lock()
	defer m.Unlock()

	if _, ok := m.Clients[c]; ok {
		c.conn.Close()
		c.leaveRoom()
		delete(m.Clients, c)
		slog.Info("client "+c.user.GetName()+" removed", fn)
		m.broadcast(UPDATE_CLIENTS_COUNTER)
	}
}

func (m *Manager) broadcast(action string) {
	fn := slog.String("func", "broadcast")

	var e event
	switch action {
	case UPDATE_CLIENTS_COUNTER:
		cc, _ := json.Marshal(len(m.Clients))
		e.Payload = cc

	case UPDATE_ROOMS:
		rooms, err := json.Marshal(m.getAllRooms())
		if err != nil {
			slog.Warn("cannot Marshal rooms", fn, "err", err)
			return
		}
		e.Payload = rooms

	default:
		slog.Warn("event had unknown action", fn, "action", action)
		return
	}

	e.Action = action
	for c := range m.Clients {
		c.writeEventBuffer <- e
	}
}

func (m *Manager) createRoom(cr CreateRoomDTO) *room {
	m.Lock()
	defer m.Unlock()

	r := newRoom(cr)
	go r.run()
	m.Rooms[r] = true
	m.broadcast(UPDATE_ROOMS)
	return r
}

func (m *Manager) removeRoom(r *room) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.Rooms[r]; ok {
		delete(m.Rooms, r)
		m.broadcast(UPDATE_ROOMS)
	}
}

func (m *Manager) findRoomById(id uuid.UUID) *room {
	for r := range m.Rooms {
		if r.Id == id {
			return r
		}
	}
	return nil
}

func (m *Manager) getAllRooms() (rooms []*room) {
	for r := range m.Rooms {
		rooms = append(rooms, r)
	}
	return
}
