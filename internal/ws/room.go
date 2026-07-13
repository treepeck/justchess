package ws

import (
	"justchess/internal/db"
	"log"
)

// room is a WebSocket endpoint (game or matchmaking queue).
// Single [client] can join the same [room] multiple times.
type room interface {
	register(p db.Player, c *client)
	unregister(c *client)
}

type queueRoom struct {
	clients map[*client]db.Player
}

func initQueueRoom() queueRoom {
	return queueRoom{
		clients: make(map[*client]db.Player),
	}
}

func (r queueRoom) register(p db.Player, c *client) {
	log.Printf("")
	r.clients[c] = p
}

func (r queueRoom) unregister(c *client) {
	delete(r.clients, c)
}

type gameRoom struct {
	clients map[*client]db.Player
}

func initGameRoom() gameRoom {
	return gameRoom{
		clients: make(map[*client]db.Player),
	}
}

func (r gameRoom) register(p db.Player, c *client) {
	r.clients[c] = p
}

func (r gameRoom) unregister(c *client) {
	delete(r.clients, c)
}
