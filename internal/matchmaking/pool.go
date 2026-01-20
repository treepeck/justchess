package matchmaking

import (
	"log"
	"math"
)

const (
	defaultThreshold = 500.0
	maxThreshold     = 3000.0
)

// Pool wraps a single Red-Black Tree and provides implementation of the
// matchmaking algorithm.
type Pool struct {
	tree *redBlackTree
}

func NewPool() Pool { return Pool{tree: newRedBlackTree()} }

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

func (p Pool) Size() int {
	return p.tree.size
}

// MakeMatches finds best matches between all players in the pool. Every found
// match is send to a result channel.
// The results channel will be closed after algorithm finishes.
func (p Pool) MakeMatches(results chan<- [2]string) {
	p.makeMatches(p.tree.root, results)
	close(results)
}

func (p Pool) makeMatches(n *redBlackNode, results chan<- [2]string) {
	if n == p.tree.leaf {
		return
	}

	// Find possible matches.
	matches := [4]*redBlackNode{n.left, n.right, p.tree.leaf, p.tree.leaf}
	if n.left != p.tree.leaf {
		matches[2] = p.tree.findMax(n.left)
	}
	if n.right != p.tree.leaf {
		matches[3] = p.tree.findMin(n.right)
	}

	// Find the match which has the lowest rating gap.
	var best *redBlackNode
	bestGap := -1.0
	for _, match := range matches {
		// Skip possible leaf nodes.
		if match == p.tree.leaf {
			continue
		}

		gap := math.Abs(n.key.rating - match.key.rating)
		if bestGap == -1.0 {
			best = match
			bestGap = gap
			continue
		}

		if gap < bestGap {
			best = match
			bestGap = gap
		}
	}

	if best == nil {
		return
	}

	// Check does the lowest gap exceeds the alowed threshold.
	if bestGap <= n.key.threshold && bestGap <= best.key.threshold {
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

// ExpandThresholds expands rating threshold of each thee node so that players
// with greater rating gaps can be paired together.
func (p Pool) ExpandThresholds() {
	if p.tree.size < 1 {
		return
	}
	p.expandThresholds(p.tree.root)
}

func (p Pool) expandThresholds(n *redBlackNode) {
	if n.key.threshold < maxThreshold {
		n.key.threshold += defaultThreshold
	}

	if n.left != p.tree.leaf {
		p.expandThresholds(n.left)
	}

	if n.right != p.tree.leaf {
		p.expandThresholds(n.right)
	}
}
