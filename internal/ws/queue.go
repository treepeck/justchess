package ws

import (
	"justchess/internal/matchmaking"
	"justchess/internal/randgen"
	"justchess/internal/talk"
	"log"
	"math/rand/v2"
)

const msgGameCreationFailed = "Please reload the page to restore the connection"

type queue struct {
	clients     map[string]*client
	register    chan registrant
	unregister  chan string
	create      chan talk.GameCreator
	pool        *matchmaking.Pool
	timeControl int
	timeBonus   int
}

func initQueue(c, b int, create chan talk.GameCreator) queue {
	q := queue{
		clients:     make(map[string]*client),
		register:    make(chan registrant),
		unregister:  make(chan string),
		create:      create,
		pool:        matchmaking.NewPool(),
		timeControl: c,
		timeBonus:   b,
	}
	go q.listen()
	return q
}

func (q queue) listen() {
	for {
		select {
		case r := <-q.register:
			q.onRegister(r)
		case id := <-q.unregister:
			q.onUnregister(id)
		case <-q.pool.Ticker.C:
			for ids := range q.pool.Matchmaking() {
				q.onMatch(ids)
			}
			q.pool.ExpandMMRGaps()
		}
	}
}

func (q queue) onRegister(r registrant) {
	if _, exists := q.clients[r.id]; exists || len(q.clients) == clientsThreshold {
		r.err <- errAlreadyRegistered
		return
	}
	defer func() {
		r.err <- nil
	}()

	conn, err := upgrader.Upgrade(r.res, r.req, nil)
	if err != nil {
		// Upgrader writes the response, so simply return here.
		return
	}
	c := newClient(conn, nil, q.unregister)
	go c.read(r.id)
	go c.write()
	q.clients[r.id] = c

	msg, err := talk.JSON(talk.MessageClientsCounter, len(q.clients))
	if err != nil {
		log.Print(err)
		return
	}
	q.broadcast(msg)
}

func (q queue) onUnregister(id string) {
	if _, exists := q.clients[id]; !exists {
		return
	}
	delete(q.clients, id)
	msg, err := talk.JSON(talk.MessageClientsCounter, len(q.clients))
	if err != nil {
		log.Print(err)
		return
	}
	q.broadcast(msg)
}

func (q queue) onMatch(ids [2]string) {
	// Randomly select players' sides.
	whiteId, blackId := ids[0], ids[1]
	if rand.IntN(2) == 1 {
		whiteId = ids[1]
		blackId = ids[0]
	}

	// Discard game creation if at least one of the players is not online.
	w := q.clients[whiteId]
	b := q.clients[blackId]
	// Notify connected clients about error.
	if w == nil || b == nil {
		msg, _ := talk.JSON(talk.MessageError, msgGameCreationFailed)
		if w != nil {
			w.send <- msg
		}
		if b != nil {
			b.send <- msg
		}
		return
	}

	gameId := randgen.GenId(randgen.IdLen)
	gc := talk.GameCreator{
		Id:          gameId,
		WhiteId:     whiteId,
		BlackId:     blackId,
		TimeControl: q.timeControl,
		TimeBonus:   q.timeBonus,
		Res:         make(chan error),
	}
	q.create <- gc

	var msg []byte
	if err := <-gc.Res; err != nil {
		msg, _ = talk.JSON(talk.MessageError, msgGameCreationFailed)
	} else {
		msg, _ = talk.JSON(talk.MessageRedirect, "/rated/"+gameId)
	}
	w.send <- msg
	b.send <- msg
}

// broadcast sends msg to all connected clients.
func (q queue) broadcast(msg []byte) {
	for _, c := range q.clients {
		c.send <- msg
	}
}
