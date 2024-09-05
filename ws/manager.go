package ws

import (
	"encoding/json"
	"log"
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
	clients map[*client]bool
	rooms   map[*room]bool
}

// Creates a new Manager.
func NewManager() *Manager {
	return &Manager{
		clients: make(map[*client]bool),
		rooms:   make(map[*room]bool),
	}
}

// Upgrades the incoming HTTP connection to the WebSocket Protocol.
// If the connection cannot be upgraded, sends a header with status code 500
// back to the client.
func (m *Manager) HandleConnection(rw http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Println("HandleConnection: error while upgrading the connection: ", err)
		return
	}

	c := newClient(conn, m)
	m.addClient(c)

}

// Adds a new client to the clients map and invokes the client`s goroutines:
//  1. readEvents goroutine handles the incomming events from the client;
//  2. writeEvent goroutine grabs the events from the evBuf channel and sends those
//     events to the client.
func (m *Manager) addClient(c *client) {
	m.Lock()
	defer m.Unlock()

	m.clients[c] = true
	log.Println("addClient: current clients count: ", len(m.clients))

	go c.readEvents()
	go c.writeEvents()

	m.broadcast(UPDATE_CLIENTS_COUNTER)
}

// Removes client from the clients map. Closes a connection with the front-ent.
func (m *Manager) removeClient(c *client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[c]; ok {
		c.conn.Close()
		c.leaveRoom()
		delete(m.clients, c)
		log.Println("removeClient: connection closed, clients count: ", len(m.clients))
		m.broadcast(UPDATE_CLIENTS_COUNTER)
	}
}

func (m *Manager) broadcast(action string) {
	var e event
	switch action {
	case UPDATE_CLIENTS_COUNTER:
		cc, _ := json.Marshal(len(m.clients))
		e.Payload = cc

	case UPDATE_ROOMS:
		rooms, err := json.Marshal(m.getAllRooms())
		if err != nil {
			log.Println("broadcast: cannot Marshal rooms ", err)
			return
		}
		e.Payload = rooms

	default:
		log.Println("broadcast: event had unknown action ", action)
		return
	}

	e.Action = action
	for c := range m.clients {
		c.writeEventBuffer <- e
	}
}

func (m *Manager) createRoom() *room {
	m.Lock()
	defer m.Unlock()

	r := newRoom()
	go r.run()
	m.rooms[r] = true
	m.broadcast(UPDATE_ROOMS)
	return r
}

func (m *Manager) findRoomById(id uuid.UUID) *room {
	for r := range m.rooms {
		if r.Id == id {
			return r
		}
	}
	return nil
}

func (m *Manager) getAllRooms() (rooms []*room) {
	for r := range m.rooms {
		rooms = append(rooms, r)
	}
	return
}
