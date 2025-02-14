package ws

import (
	"encoding/binary"
	"log"
	"net/http"

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

type Manager struct {
	// Use of empty struct is an optimization.
	clients    map[*client]struct{}
	rooms      map[*room]struct{}
	register   chan *client
	unregister chan *client
	add        chan *room
	remove     chan *room
}

func NewManager() *Manager {
	return &Manager{
		clients:    make(map[*client]struct{}),
		rooms:      make(map[*room]struct{}),
		register:   make(chan *client),
		unregister: make(chan *client),
		add:        make(chan *room),
		remove:     make(chan *room),
	}
}

func (m *Manager) HandleNewConnection(rw http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	c := newClient(conn, m)
	m.register <- c
}

func (m *Manager) Run() {
	for {
		select {
		case c, ok := <-m.register:
			if !ok {
				return
			}
			m.addClient(c)
			msg := make([]byte, 5)
			l := len(m.clients)
			msg[0] = uint8(l) & 0xF
			msg[1] = uint8(l>>8) & 0xF
			msg[2] = uint8(l>>16) & 0xF
			msg[3] = uint8(l>>24) & 0xF
			msg[4] = CLIENTS_COUNTER
			m.broadcast(msg)
			// Send availible rooms for the client.
			msg = make([]byte, 19)
			for r := range c.manager.rooms {
				copy(msg[0:16], r.id[0:16])
				msg[16] = r.game.TimeControl
				msg[17] = r.game.TimeBonus
				msg[18] = ADD_ROOM
				c.send <- msg
			}

		case c := <-m.unregister:
			if _, ok := m.clients[c]; ok {
				m.removeClient(c)
				msg := make([]byte, 5)
				binary.LittleEndian.PutUint32(msg, uint32(len(m.clients)))
				msg[4] = CLIENTS_COUNTER
				m.broadcast(msg)
			}

		case r := <-m.add:
			m.addRoom(r)
			msg := make([]byte, 19)
			copy(msg[0:16], r.id[0:16])
			msg[16] = r.game.TimeControl
			msg[17] = r.game.TimeBonus
			msg[18] = ADD_ROOM
			m.broadcast(msg)

		case r := <-m.remove:
			m.removeRoom(r)
			msg := make([]byte, 17)
			copy(msg[0:15], r.id[:])
			msg[16] = REMOVE_ROOM
			m.broadcast(msg)
		}
	}
}

func (m *Manager) addClient(c *client) {
	// If the client is alredy connected, close the previous connection.
	for connectedC := range m.clients {
		if connectedC.id == c.id {
			m.removeClient(connectedC)
		}
	}
	m.clients[c] = struct{}{}
	log.Printf("client %s added\n", c.id.String())

	go c.readPump()
	go c.writePump()
}

func (m *Manager) removeClient(c *client) {
	delete(m.clients, c)
	close(c.send)
	log.Printf("client %s removed\n", c.id.String())
}

func (m *Manager) addRoom(r *room) {
	m.rooms[r] = struct{}{}
	log.Printf("room %s added\n", r.id.String())
}

func (m *Manager) removeRoom(r *room) {
	delete(m.rooms, r)
	log.Printf("room %s removed\n", r.id.String())
}

// broadcast the specified message among all connected clients.
func (m *Manager) broadcast(msg []byte) {
	for c := range m.clients {
		// Send messages only to clients that are currently not in the game.
		if c.currentRoom == nil {
			c.send <- msg
		}
	}
}
