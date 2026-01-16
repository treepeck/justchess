package matchmaking

import (
	"log"
)

type Pool struct {
	tree *redBlackTree
}

func NewPool() Pool {
	return Pool{
		tree: newRedBlackTree(),
	}
}

// Join associates the specified [Ticket] with the player's id and their rating
// and returns the number of players who have selected the same [Ticket].
// It's the caller's responsibility to ensure that a single client doesn't join
// more than once.
func (p Pool) Join(id string, rating float64) {
	n := p.tree.spawn(rating, id)
	p.tree.insertNode(n)
}

// Leave removes the record with specified id from the tickets map and deletes
// the player's node from the tree and returns the number of active players
// who have selected the same ticket.
func (p Pool) Leave(id string, rating float64) {
	n := p.tree.search(rating, id)
	if n == nil {
		return
	}
	p.tree.removeNode(n)
	log.Printf("Player %s leaves matchmaking", id)
}

func (p Pool) MakeMatches() {

}
