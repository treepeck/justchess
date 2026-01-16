package ws

import (
	"encoding/json"
	"justchess/internal/matchmaking"
	"log"
)

type queue struct {
	pool       matchmaking.Pool
	register   chan handshake
	unregister chan *client
	// TODO: store wait time instead of empty struct.
	clients map[string]*client
	// Matchmaking parameters.
	control int
	bonus   int
}

func newQueue(control, bonus int) queue {
	return queue{
		pool:       matchmaking.NewPool(),
		register:   make(chan handshake),
		unregister: make(chan *client),
		clients:    make(map[string]*client),
		control:    control,
		bonus:      bonus,
	}
}

func (q queue) listenEvents() {
	for {
		select {
		case c := <-q.register:
			q.handleRegister(c)
			q.broadcastClientsCounter()

		case c := <-q.unregister:
			q.handleUnregister(c)
			q.broadcastClientsCounter()
		}
	}
}

func (q queue) handleRegister(h handshake) {
	// Write to the response channel so that request cannot be closed.
	defer func() { h.ch <- struct{}{} }()

	// Deny the request if the clients is already in the queue.
	if _, exists := q.clients[h.player.Id]; exists {
		return
	}

	conn, err := upgrader.Upgrade(h.rw, h.r, nil)
	if err != nil {
		// upgrader writes the response, so simply return here.
		return
	}

	c := newClient(h.player, conn)
	go c.read(q.unregister, nil)
	go c.write()

	q.pool.Join(c.player.Id, c.player.Rating)
	log.Printf("client %s joined queue", c.player.Id)
}

func (q queue) handleUnregister(c *client) {
	if _, exists := q.clients[c.player.Id]; !exists {
		log.Printf("client %s is not registered", c.player.Id)
		return
	}

	delete(q.clients, c.player.Id)
	q.pool.Leave(c.player.Id, c.player.Rating)
	log.Printf("client %s leaved queue", c.player.Id)
}

// broadcast clients counter event among all connected clients.
func (q queue) broadcastClientsCounter() {
	// Encode event payload.
	raw, err := json.Marshal(len(q.clients))
	if err != nil {
		log.Print(err)
		return
	}

	e, err := json.Marshal(event{Action: actionClientsCounter, Payload: raw})
	if err != nil {
		log.Print(err)
		return
	}

	for _, c := range q.clients {
		c.send <- e
	}
}
