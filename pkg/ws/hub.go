package ws

import (
	"encoding/json"
	"log"
)

// Hub stores all connected clients and created rooms.
//
// Each client message (event) comes into the bus channel, which
// will call the corresponding event handler.
type Hub struct {
	// Connected clients which are subscribed to the hub events.
	subs       map[string]*client
	rooms      map[string]*room
	register   chan *client
	unregister chan *client
	bus        chan event
	// Number of currently connected clients.
	counter uint
}

func NewHub() *Hub {
	h := &Hub{
		register:   make(chan *client),
		unregister: make(chan *client),
		bus:        make(chan event),
		subs:       make(map[string]*client),
		rooms:      make(map[string]*room),
	}
	go h.route()
	return h
}

// route consequentially (one at a time) extracts events and calls
// the corresponding event handler.
func (h *Hub) route() {
	for {
		select {
		case c := <-h.register:
			h.placeClient(c)

		case c := <-h.unregister:
			h.removeClient(c)

		case e := <-h.bus:
			h.handle(e)
		}
	}
}

// placeClient registers a new subscriber to a specific room or to the hub.
//
// The registration process includes the following steps:
//  1. Check the room id provided in the query, which indicates
//     the room the client wants to connect to.
//  2. If the room id is empty or no room with that id exists,
//     subscribe the client to hub events.
//  3. If a room with the specified id exists, subscribe the client
//     to that room.
//  4. Increment the counter.
//  5. Publish the counter value to all hub subscribers.
func (h *Hub) placeClient(c *client) {
	if r := h.rooms[c.subscribtionId]; r != nil {
		r.register(c)
	} else {
		c.subscribtionId = ""
		h.subs[c.id] = c
	}

	h.counter++
	h.publish(event{Action: actionCounter, Payload: encode(h.counter)})
}

// removeClient unregisters the subscriber from the topic to which it is subscribed.
func (h *Hub) removeClient(c *client) {
	if c.subscribtionId == "" {
		delete(h.subs, c.id)
	} else {
		r := h.rooms[c.subscribtionId]
		r.unregister(c.id)

		// Remove room is there are no connected subscribers left.
		if len(r.subs) == 0 {
			h.removeRoom(r.id)
		}
	}

	h.counter--
	h.publish(event{Action: actionCounter, Payload: encode(h.counter)})
}

// addRoom inserts a new room into rooms and broadcasts the room among all hub subscribers.
func (h *Hub) addRoom() {
	r := newRoom()
	h.rooms[r.id] = r
	h.publish(event{Action: actionCreate, Payload: encode(r.id)})
}

func (h *Hub) removeRoom(rid string) {
	delete(h.rooms, rid)

	h.publish(event{Action: actionRemove, Payload: encode(rid)})
}

func (h *Hub) handle(e event) {
	switch e.Action {
	case actionCreate:
		h.addRoom()

	}
}

func (h *Hub) publish(e event) {
	for _, c := range h.subs {
		c.send <- e
	}
}

func encode(payload any) []byte {
	p, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ERROR: cannot encode payload %v", err)
	}
	return p
}
