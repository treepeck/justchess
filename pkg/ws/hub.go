package ws

import (
	"encoding/json"
	"justchess/pkg/auth"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// upgrader is used to upgrate the HTTP connection into the websocket protocol.
var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin:     func(r *http.Request) bool { return r.Header.Get("Origin") == "http://localhost:3000" },
}

// Hub is a global repository of all created rooms and connected clients which are not in the game.
// To ensure safe concurrent access, the Hub is protected with a Mutex.
type Hub struct {
	sync.Mutex
	clients map[*client]struct{}
	rooms   map[*Room]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*client]struct{}),
		rooms:   make(map[*Room]struct{}),
	}
}

func (h *Hub) HandleNewConnection(rw http.ResponseWriter, r *http.Request) {
	s := r.Context().Value(auth.Subj)
	if s == nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	subj := s.(auth.Subject)

	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Printf("request from: %s\n", r.Header.Get("Origin"))
		log.Printf("%v\n", err)
		return
	}

	c := newClient(subj.Id, subj.Name, subj.Role == auth.RoleGuest, conn)
	c.hub = h

	go c.readRoutine()
	go c.writeRoutine()

	h.register(c)
}

// register denies multiple connections from a single peer.
func (h *Hub) register(c *client) {
	h.Lock()
	defer h.Unlock()

	for connected := range h.clients {
		if connected.id == c.id {
			return
		}
	}

	h.clients[c] = struct{}{}
	log.Printf("client %s registered\n", c.id.String())

	h.broadcastClientsCounter()
	h.send10Rooms(c)
}

// unregister removes the client if it is connected.
func (h *Hub) unregister(c *client) {
	h.Lock()
	defer h.Unlock()

	if _, ok := h.clients[c]; !ok {
		return
	}

	delete(h.clients, c)
	log.Printf("client %s unregistered\n", c.id.String())

	h.broadcastClientsCounter()
}

// add denies multiple room creation.
func (h *Hub) add(r *Room) {
	h.Lock()
	defer h.Unlock()

	h.rooms[r] = struct{}{}
	log.Printf("room %s added\n", r.id.String())

	h.broadcastAddRoom(r)
}

// remove terminates the room's handleMessages routine.
func (h *Hub) remove(r *Room) {
	h.Lock()
	defer h.Unlock()

	if _, ok := h.rooms[r]; !ok {
		return
	}

	delete(h.rooms, r)
	log.Printf("room %s removed\n", r.id.String())

	h.broadcastRemoveRoom(r.id)
}

// broadcastClientsCounter does not Lock the hub, so it cannot be called in a non-blocking routine!
func (h *Hub) broadcastClientsCounter() {
	data, err := json.Marshal(ClientsCounterData{Counter: len(h.clients)})
	if err != nil {
		log.Printf("cannot Marshal message: %v\n", err)
		return
	}

	msg, _ := json.Marshal(Message{Type: CLIENTS_COUNTER, Data: data})

	for c := range h.clients {
		c.send <- msg
	}
}

// broadcastAddRoom does not Lock the hub, so it cannot be called in a non-blocking routine!
// Room's game field should not be nil!
func (h *Hub) broadcastAddRoom(r *Room) {
	data, err := json.Marshal(AddRoomData{
		Id:          r.id,
		Creator:     r.creatorName,
		Players:     r.getClientIds(),
		TimeControl: r.game.TimeControl,
		TimeBonus:   r.game.TimeBonus,
	})
	if err != nil {
		log.Printf("cannot Marshal message: %v\n", err)
		return
	}

	msg, _ := json.Marshal(Message{Type: ADD_ROOM, Data: data})

	for c := range h.clients {
		c.send <- msg
	}
}

func (h *Hub) broadcastRemoveRoom(roomId uuid.UUID) {
	data, err := json.Marshal(RemoveRoomData{
		RoomId: roomId.String(),
	})
	if err != nil {
		log.Printf("cannot Marshal message: %v\n", err)
		return
	}

	msg, _ := json.Marshal(Message{Type: REMOVE_ROOM, Data: data})

	for c := range h.clients {
		c.send <- msg
	}
}

// send10Rooms does not Lock the hub, so it cannot be called in a non-blocking routine!
// Each room's game field should not be nil!
func (h *Hub) send10Rooms(c *client) {
	cnt := 0

	for r := range h.rooms {
		cnt++
		if cnt == 10 {
			return
		}

		data, err := json.Marshal(AddRoomData{
			Id:          r.id,
			Creator:     r.creatorName,
			Players:     r.getClientIds(),
			TimeControl: r.game.TimeControl,
			TimeBonus:   r.game.TimeBonus,
		})
		if err != nil {
			log.Printf("cannot Marshal message: %v\n", err)
			return
		}

		msg, _ := json.Marshal(Message{Type: ADD_ROOM, Data: data})

		c.send <- msg
	}
}

func (h *Hub) GetRoomById(id uuid.UUID) *Room {
	h.Lock()
	defer h.Unlock()

	for r := range h.rooms {
		if r.id == id {
			return r
		}
	}
	return nil
}
