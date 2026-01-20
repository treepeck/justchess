package ws

import (
	"encoding/json"
	"justchess/internal/matchmaking"
	"justchess/internal/randgen"
	"log"
	"time"
)

const matchmakingTick = 5 * time.Second

type queue struct {
	ticker     *time.Ticker
	pool       matchmaking.Pool
	register   chan handshake
	unregister chan *client
	clients    map[string]*client
	// Matchmaking parameters.
	control int
	bonus   int
}

func newQueue(control, bonus int) queue {
	return queue{
		ticker:     time.NewTicker(matchmakingTick),
		pool:       matchmaking.NewPool(),
		register:   make(chan handshake),
		unregister: make(chan *client),
		clients:    make(map[string]*client),
		control:    control,
		bonus:      bonus,
	}
}

func (q queue) listenEvents(add chan addRoomEvent) {
	for {
		select {
		case c := <-q.register:
			q.handleRegister(c)
			q.broadcastClientsCounter()

		case c := <-q.unregister:
			q.handleUnregister(c)
			q.broadcastClientsCounter()

		case <-q.ticker.C:
			matches := make(chan [2]string)
			go q.pool.MakeMatches(matches)

			for {
				match, ok := <-matches
				if !ok {
					break
				}
				roomId := randgen.GenId(randgen.IdLen)
				// Add room to service.
				add <- addRoomEvent{
					playerIDs: match,
					roomId:    roomId,
					control:   q.control,
					bonus:     q.bonus,
				}
				// Notify clients.
				q.sendRedirect(match, roomId)
			}
		}
	}
}

func (q queue) handleRegister(h handshake) {
	// Write to the response channel so that request cannot be closed.
	defer func() { h.ch <- struct{}{} }()

	// Deny the request if the client is already in the queue.
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

	q.clients[h.player.Id] = c
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

func (q queue) sendRedirect(players [2]string, roomId string) {
	// Encode event payload.
	raw, err := json.Marshal(roomId)
	if err != nil {
		log.Print(err)
		return
	}

	e, err := json.Marshal(event{Action: actionRedirect, Payload: raw})
	if err != nil {
		log.Print(err)
		return
	}

	for _, c := range q.clients {
		if c.player.Id == players[0] || c.player.Id == players[1] {
			c.send <- e
		}
	}
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
