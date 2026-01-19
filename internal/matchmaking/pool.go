package matchmaking

import (
	"log"
	"math"
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

// MakeMatches finds best matches between all players in the pool. Every found
// match is send to a result channel.
// It's the caller's responsibility to close the result channel after this function
// exists.
func (p Pool) MakeMatches(n *redBlackNode, result chan<- [2]string) {
	// Shortcut: node doesn't have any matches.
	if n == p.tree.leaf || n.left == p.tree.leaf || n.right == p.tree.leaf {
		return
	}

	// Find possible matches.
	matches := [4]*redBlackNode{
		n.left, n.right, p.tree.findMax(n.left), p.tree.findMin(n.right),
	}

	// Find the match, which has the lowest rating gap.
	best := matches[0]
	lowestGap := math.Abs(n.key.rating - best.key.rating)
	for i := 1; i < len(matches); i++ {
		gap := math.Abs(n.key.rating - matches[i].key.rating)
		if gap < lowestGap {
			best = matches[i]
			lowestGap = gap
		}
	}

	// Check does the lowest gap exceeds the alowed threshold.
	if lowestGap <= n.key.gapThreshold && lowestGap <= best.key.gapThreshold {
		// Notify service about created rooms.
		result <- [2]string{n.key.playerId, best.key.playerId}

		// Remove nodes from tree.
		p.tree.removeNode(p.tree.search(n.key.rating, n.key.playerId))
		p.tree.removeNode(p.tree.search(best.key.rating, best.key.playerId))

		// Call function recursively.
		p.MakeMatches(p.tree.root, result)
		return
	}

	// Call function recursively on left and right subtrees.
	if n.left != p.tree.leaf {
		p.MakeMatches(n.left, result)
	}

	if n.right != p.tree.leaf {
		p.MakeMatches(n.right, result)
	}
}
