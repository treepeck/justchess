package matchmaking

import (
	"iter"
	"log"
	"math"
)

const (
	defaultMaxGap = 500.0
	gapLimit      = 3000.0
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

// MakeMatches finds best matches between all players in the pool. Every found
// match is send to a result channel.
// The results channel will be closed after algorithm finishes.
func (p Pool) MakeMatches() iter.Seq[[2]string] {
	n := p.tree.root

	return func(yield func([2]string) bool) {
		p.makeMatches(n, yield)
	}
}

func (p Pool) makeMatches(n *redBlackNode, yield func([2]string) bool) {
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
	bestGap := gapLimit
	for _, match := range matches {
		// Skip leaf and duplicate nodes. A single client is allowed to join
		// multiple times.
		if match == p.tree.leaf || match.key.playerId == n.key.playerId {
			continue
		}

		gap := math.Abs(n.key.rating - match.key.rating)
		if gap < bestGap {
			bestGap = gap
			best = match
		}
	}

	if best == nil {
		return
	}

	// Check does the best gap exceeds the allowed rating gap.
	if bestGap <= n.key.maxGap && bestGap <= best.key.maxGap {
		if !yield([2]string{n.key.playerId, best.key.playerId}) {
			return
		}

		// Remove matched nodes from tree.
		p.tree.removeNode(p.tree.search(n.key.rating, n.key.playerId))
		p.tree.removeNode(p.tree.search(best.key.rating, best.key.playerId))

		// Call function recursively.
		p.makeMatches(p.tree.root, yield)
		return
	}

	// Call function recursively on left and right subtrees.
	if n.left != p.tree.leaf {
		p.makeMatches(n.left, yield)
	}

	if n.right != p.tree.leaf {
		p.makeMatches(n.right, yield)
	}
}

// ExpandRatingGaps expands the allowed rating gap of each player so that players
// with larger rating gaps can eventually be paired together.
func (p Pool) ExpandRatingGaps() {
	if p.tree.size < 1 {
		return
	}
	p.expandThresholds(p.tree.root)
}

func (p Pool) expandThresholds(n *redBlackNode) {
	if n.key.maxGap < gapLimit {
		n.key.maxGap += defaultMaxGap
	}

	if n.left != p.tree.leaf {
		p.expandThresholds(n.left)
	}

	if n.right != p.tree.leaf {
		p.expandThresholds(n.right)
	}
}
