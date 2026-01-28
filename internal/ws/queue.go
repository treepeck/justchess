package ws

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"time"

	"justchess/internal/db"
	"justchess/internal/matchmaking"
	"justchess/internal/randgen"
)

const (
	// Declaration of error messages.
	msgRoomCreationFailed = "Please reload the page to restore the connection"

	// Interval at which the matchmaking process will occur.
	matchmakingTick = 3 * time.Second
)

type queue struct {
	ticker     *time.Ticker
	pool       matchmaking.Pool
	register   chan handshake
	unregister chan *client
	clients    map[*client]db.Player
	// Matchmaking parameters.
	control int
	bonus   int
}

func newQueue(control int, bonus int) queue {
	return queue{
		ticker:     time.NewTicker(matchmakingTick),
		pool:       matchmaking.NewPool(),
		register:   make(chan handshake),
		unregister: make(chan *client),
		clients:    make(map[*client]db.Player),
		control:    control,
		bonus:      bonus,
	}
}

// listenEvent handles concurrent client registration, unregistration and
// matchmaking ticks. create chan is used to notify the service about new
// matched.
func (q queue) listenEvents(create chan<- createRoomEvent) {
	for {
		select {
		case c := <-q.register:
			q.handleRegister(c)
			q.broadcastClientsCounter()

		case c := <-q.unregister:
			q.handleUnregister(c)
			q.broadcastClientsCounter()

		case <-q.ticker.C:
			for match := range q.pool.MakeMatches() {
				q.handleMatch(match, create)
			}
			q.pool.ExpandRatingGaps()
		}
	}
}

func (q queue) handleRegister(h handshake) {
	// Write to the response channel so that request cannot be closed.
	defer func() { h.ch <- struct{}{} }()

	conn, err := upgrader.Upgrade(h.rw, h.r, nil)
	if err != nil {
		// upgrader writes the response, so simply return here.
		return
	}

	c := newClient(conn)
	go c.read(q.unregister, nil)
	go c.write()

	q.clients[c] = h.player

	// Join the matchmaking pool.
	q.pool.Join(h.player.Id, h.player.Rating)
	log.Printf("client %s joined queue", h.player.Id)
}

func (q queue) handleUnregister(c *client) {
	p, exists := q.clients[c]
	if !exists {
		log.Printf("client is not registered")
		return
	}

	delete(q.clients, c)
	q.pool.Leave(p.Id, p.Rating)

	log.Printf("client %s leaved queue", p.Id)
}

func (q queue) handleMatch(match [2]string, create chan<- createRoomEvent) {
	roomId := randgen.GenId(randgen.IdLen)

	// Randomly select players' sides.
	whiteId, blackId := match[0], match[1]
	if rand.IntN(2) == 1 {
		whiteId = match[1]
		blackId = match[0]
	}

	// Send create room event.
	e := createRoomEvent{
		id: roomId, whiteId: whiteId, blackId: blackId,
		control: q.control, bonus: q.bonus, res: make(chan error, 1),
	}
	create <- e

	// Wait for the response.
	err := <-e.res
	if err != nil {
		// Notify clients about error.
		q.sendEvent(match, actionError, msgRoomCreationFailed)
	} else {
		// Redirect clients to game room.
		q.sendEvent(match, actionRedirect, roomId)
	}
}

func (q queue) sendEvent(players [2]string, a eventAction, payload string) {
	// Encode event payload.
	raw, err := json.Marshal(payload)
	if err != nil {
		log.Print(err)
		return
	}

	e, err := json.Marshal(event{Action: a, Payload: raw})
	if err != nil {
		log.Print(err)
		return
	}

	for c, p := range q.clients {
		if p.Id == players[0] || p.Id == players[1] {
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

	for c := range q.clients {
		c.send <- e
	}
}
