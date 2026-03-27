package ws

import (
	"log"
	"math/rand/v2"
	"time"

	"justchess/internal/db"
	"justchess/internal/event"
	"justchess/internal/game"
	"justchess/internal/mm"
	"justchess/internal/randgen"
)

const (
	// Declaration of error messages.
	msgRoomCreationFailed = "Please reload the page to restore the connection"

	// Interval at which the matchmaking process will occur.
	matchmakingTick = 3 * time.Second
)

type queue struct {
	gameRepo   db.GameRepo
	playerRepo db.PlayerRepo
	clients    map[string]*client
	pool       mm.Pool
	create     chan createRoomPayload
	register   chan *client
	unregister chan string
	ticker     *time.Ticker
	// Matchmaking parameters.
	control int // In seconds.
	bonus   int // In seconds.
}

func newQueue(create chan createRoomPayload,
	control, bonus int,
	gr db.GameRepo, pr db.PlayerRepo,
) queue {
	return queue{
		gameRepo:   gr,
		playerRepo: pr,
		ticker:     time.NewTicker(matchmakingTick),
		pool:       mm.NewPool(),
		create:     create,
		register:   make(chan *client),
		unregister: make(chan string),
		clients:    make(map[string]*client),
		control:    control,
		bonus:      bonus,
	}
}

// listenEvent handles concurrent client registration, unregistration and
// matchmaking ticks.
func (q queue) listenEvents() {
	for {
		select {
		case c := <-q.register:
			q.add(c)
			q.broadcast(event.JSON(event.ClientsCounter, len(q.clients)))

		case id := <-q.unregister:
			q.remove(id)
			q.broadcast(event.JSON(event.ClientsCounter, len(q.clients)))

		case <-q.ticker.C:
			for ids := range q.pool.MakeMatches() {
				q.match(ids)
			}
			q.pool.ExpandRatingGaps()
		}
	}
}

func (q queue) add(c *client) {
	if len(q.clients) == clientsThreshold {
		c.send <- event.JSON(event.Error, msgTooMany)
		return
	}

	if _, exist := q.clients[c.player.Id]; exist {
		// Send error event to the client.
		c.send <- event.JSON(event.Error, msgConflict)
		return
	}

	if c.player.IsGuest {
		// Redirect guest players to signup page.
		c.send <- event.JSON(event.Redirect, "/signup")
		return
	}

	c.unregister = q.unregister
	q.clients[c.player.Id] = c
	// Join the matchmaking pool.
	q.pool.Join(c.player.Id, c.player.Rating)
}

func (q queue) remove(id string) {
	c, exist := q.clients[id]
	if !exist {
		log.Printf("client %s is not registered", id)
		return
	}
	delete(q.clients, id)
	q.pool.Leave(id, c.player.Rating)
}

func (q queue) match(ids [2]string) {
	roomId := randgen.GenId(randgen.IdLen)

	// Randomly select players' sides.
	whiteId, blackId := ids[0], ids[1]
	if rand.IntN(2) == 1 {
		whiteId = ids[1]
		blackId = ids[0]
	}

	// If player's are not online, cancel.
	w := q.clients[whiteId]
	b := q.clients[blackId]
	if w == nil || b == nil {
		// Notify clients about error.
		q.sendEvent(ids, event.JSON(event.Error, msgRoomCreationFailed))
		return
	}

	g, err := game.SpawnRatedGame(
		w.player, b.player, q.control, q.bonus,
		roomId, q.gameRepo, q.playerRepo,
	)
	if err != nil {
		// Notify clients about error.
		q.sendEvent(ids, event.JSON(event.Error, msgRoomCreationFailed))
	} else {
		p := createRoomPayload{
			id:   roomId,
			game: g,
			res:  make(chan struct{}, 1),
		}
		q.create <- p
		// Wait for response to redirect clients only after room is ready.
		<-p.res

		// Redirect clients to room.
		q.sendEvent(ids, event.JSON(event.Redirect, "/rated/"+roomId))
	}
}

func (q queue) sendEvent(players [2]string, raw []byte) {
	for _, id := range players {
		if c, exists := q.clients[id]; exists {
			c.send <- raw
		}
	}
}

func (q queue) broadcast(raw []byte) {
	for _, c := range q.clients {
		c.send <- raw
	}
}
