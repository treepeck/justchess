package ws

import (
	"encoding/json"
	"justchess/pkg/auth"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// upgrader is used to upgrate the HTTP connection into the websocket protocol.
var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin:     func(r *http.Request) bool { return r.Header.Get("Origin") == os.Getenv("domain") },
}

// Hub is a global repository of all created rooms and connected clients which are not in the game.
// To ensure safe concurrent access, the Hub is protected with a Mutex.
type Hub struct {
	sync.Mutex
	rooms   map[*Room]struct{}
	clients map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		rooms:   make(map[*Room]struct{}),
		clients: make(map[*client]struct{}),
	}
}

// HandleNewConnection registers a new client in the Hub. The client will recieve the
// messages about room creation and deletion and send messages about joining or creating
// the room.
func (h *Hub) HandleNewConnection(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context().Value(auth.Cms)
	if ctx == nil {
		log.Println("request with nil context")
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	cms := ctx.(auth.Claims)

	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Printf("request from: %s\n", r.Header.Get("Origin"))
		log.Printf("%v\n", err)
		return
	}

	c := newClient(cms.Id, cms.Name, cms.Role == auth.RoleGuest, conn)
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

	go r.runRoutine()

	h.rooms[r] = struct{}{}
	log.Printf("room %s added\n", r.Id.String())

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
	log.Printf("room %s removed\n", r.Id.String())

	h.broadcastRemoveRoom(r.Id)
}

// broadcastClientsCounter does not Lock the hub, so it cannot be called in a non-blocking routine!
func (h *Hub) broadcastClientsCounter() {
	data, err := json.Marshal(ClientsCounterDTO{Counter: len(h.clients)})
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
	data, err := json.Marshal(AddRoomDTO{
		Id:          r.Id,
		Creator:     r.CreatorName,
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
	data, err := json.Marshal(RemoveRoomDTO{
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

		data, err := json.Marshal(AddRoomDTO{
			Id:          r.Id,
			Creator:     r.CreatorName,
			TimeBonus:   r.game.TimeBonus,
			TimeControl: r.game.TimeControl,
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
		if r.Id == id {
			return r
		}
	}
	return nil
}
