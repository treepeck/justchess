package ws

import (
	"crypto/rand"

	"github.com/BelikovArtem/chego/game"
	"github.com/BelikovArtem/chego/movegen"
)

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
	game    *game.Game
	// Connected clients which are subscribed to the room events.
	subs map[string]*client
}

func newRoom() *room {
	return &room{
		id:   rand.Text(),
		game: game.NewGame(),
		subs: make(map[string]*client),
	}
}

func (r *room) register(c *client) {
	r.subs[c.id] = c
}

func (r *room) unregister(id string) {
	delete(r.subs, id)
}

func (r *room) handleMove(m movegen.Move) {
	moveInd := r.game.GetLegalMoveIndex(m)

	if moveInd != -1 {
		r.game.PushMove(r.game.LegalMoves.Moves[moveInd])
	}
}

func (r *room) publish(e event) {
	for _, c := range r.subs {
		c.send <- e
	}
}
