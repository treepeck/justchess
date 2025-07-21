package ws

import "crypto/rand"

// room wraps a single game and stores the clients which are subscribed
// to its events.
//
// Each room is stored in the hub's memory. A game record is inserted into the database
// only after both players have made their first moves. The initial game result is set to
// [Unscored].
//
// If both players disconnect, the room is deleted. If a player reconnects,
// the game is loaded from the database and the room is restored in memory.
//
// When the game ends, its result is updated in the database. If a player connects after
// the game has ended, no new room is created; the game information is simply displayed.
type room struct {
	// id must be equal to game.id in the database.
	id      string
	whiteId string
	blackId string
	// Connected clients which are subscribed to the room events.
	subs map[int64]*client
}

func newRoom() *room {
	return &room{
		id:   rand.Text(),
		subs: make(map[int64]*client),
	}
}

func (r *room) register(c *client) {
	r.subs[c.id] = c
}

func (r *room) unregister(id int64) {
	delete(r.subs, id)
}

func (r *room) publish(e event) {
	for _, c := range r.subs {
		c.send <- e
	}
}
