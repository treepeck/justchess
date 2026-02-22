package ws

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"time"

	"justchess/internal/mm"
	"justchess/internal/randgen"
	"justchess/internal/web"
)

const (
	// Declaration of error messages.
	msgRoomCreationFailed = "Please reload the page to restore the connection"

	// Interval at which the matchmaking process will occur.
	matchmakingTick = 3 * time.Second
)

type queue struct {
	ticker     *time.Ticker
	pool       mm.Pool
	register   chan *client
	unregister chan string
	clients    map[string]*client
	// Matchmaking parameters.
	control web.QueueData
}

func newQueue(id byte) queue {
	return queue{
		ticker:     time.NewTicker(matchmakingTick),
		pool:       mm.NewPool(),
		register:   make(chan *client),
		unregister: make(chan string),
		clients:    make(map[string]*client),
		control:    web.Controls[id-1],
	}
}

// listenEvent handles concurrent client registration, unregistration and
// matchmaking ticks. create chan is used to notify the [Service] about new
// game room.
func (q queue) listenEvents(create chan<- createRoom) {
	for {
		select {
		case c := <-q.register:
			q.handleRegister(c)
			q.broadcastClientsCounter()

		case id := <-q.unregister:
			q.handleUnregister(id)
			q.broadcastClientsCounter()

		case <-q.ticker.C:
			for match := range q.pool.MakeMatches() {
				// Handle matches.
				roomId := randgen.GenId(randgen.IdLen)

				// Randomly select players' sides.
				whiteId, blackId := match[0], match[1]
				if rand.IntN(2) == 1 {
					whiteId = match[1]
					blackId = match[0]
				}

				// Send create room event.
				e := createRoom{
					id:      roomId,
					whiteId: whiteId,
					blackId: blackId,
					control: q.control,
					res:     make(chan error, 1),
				}
				create <- e

				// Handle response.
				if <-e.res != nil {
					// Notify clients about error.
					q.sendEvent(match, actionError, msgRoomCreationFailed)
				} else {
					// Redirect clients to game room.
					q.sendEvent(match, actionRedirect, roomId)
				}
			}
			q.pool.ExpandRatingGaps()
		}
	}
}

func (q queue) handleRegister(c *client) {
	// Deny the connection if the client is already in the queue.
	if _, exist := q.clients[c.player.Id]; exist {
		// Send error event to the client.
		if raw, err := newEncodedEvent(actionError, msgConflict); err == nil {
			c.send <- raw
		} else {
			log.Print(err)
		}
		return
	}

	log.Printf("client %s joined queue", c.player.Id)

	c.unregister = q.unregister
	q.clients[c.player.Id] = c
	// Join the matchmaking pool.
	q.pool.Join(c.player.Id, c.player.Rating)
}

func (q queue) handleUnregister(id string) {
	c, exists := q.clients[id]
	if !exists {
		log.Printf("client is not registered")
		return
	}

	delete(q.clients, id)
	q.pool.Leave(c.player.Id, c.player.Rating)

	log.Printf("client %s leaved queue", id)
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

	for _, id := range players {
		if c := q.clients[id]; c != nil {
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
