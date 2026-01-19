package matchmaking

import (
	"log"
	"math"
)

// Pool wraps a single Red-Black Tree and provides implementation of the
// matchmaking algorithm.
type Pool struct {
	tree *redBlackTree
}

func NewPool() Pool {
	return Pool{
		tree: newRedBlackTree(),
	}
}

// It's the caller's responsibility to ensure that a single client doesn't join
// more than once.
func (p Pool) Join(id string, rating float64) {
	n := p.tree.spawn(rating, id)
	p.tree.insertNode(n)
}

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
// The results channel will be closed after algorithm finishes.
func (p Pool) MakeMatches(results chan<- [2]string) {
	p.makeMatches(p.tree.root, results)
	close(results)
}

func (p Pool) makeMatches(n *redBlackNode, results chan<- [2]string) {
	// Base case: node doesn't have any matches.
	if n == p.tree.leaf || n.left == p.tree.leaf || n.right == p.tree.leaf {
		return
	}

	// Find possible matches.
	matches := [4]*redBlackNode{
		n.left, n.right, p.tree.findMax(n.left), p.tree.findMin(n.right),
	}

	// Find the match which has the lowest rating gap.
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
		results <- [2]string{n.key.playerId, best.key.playerId}

		// Remove nodes from tree.
		p.tree.removeNode(p.tree.search(n.key.rating, n.key.playerId))
		p.tree.removeNode(p.tree.search(best.key.rating, best.key.playerId))

		// Call function recursively.
		p.makeMatches(p.tree.root, results)
		return
	}

	// Call function recursively on left and right subtrees.
	if n.left != p.tree.leaf {
		p.makeMatches(n.left, results)
	}

	if n.right != p.tree.leaf {
		p.makeMatches(n.right, results)
	}
}
