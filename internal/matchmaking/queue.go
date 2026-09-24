// TODO: find out how people normally implement matchmaking.
package matchmaking

import (
	"fmt"
	"iter"
	"math"
)

const (
	DefaultMaxGap float64 = 500
	GapLimit      float64 = 3000
)

// queue provides the implementation of the matchmaking algorithm.
// WARN: it's the caller's responsibility to ensure thread-safetiness.
type queue struct {
	nodes *redBlackTree
	// Number of players.
	size int
}

func newQueue() *queue {
	return &queue{
		nodes: newRedBlackTree(),
	}
}

// insert inserts a new player to the [queue] and returns it's size.
func (q *queue) insert(rating float64, id string) int {
	q.nodes.insert(q.nodes.spawn(rating, id))
	q.size++
	return q.size
}

// remove removes an existing player from the [queue] and returns it's size.
func (q *queue) remove(rating float64, id string) int {
	n := search(q.nodes.root, rating, id)
	if n == nil {
		fmt.Printf("matchmaking: trying to remove non-existing player \"%s\"\n", id)
		return q.size
	}
	q.nodes.remove(n)
	q.size--
	return q.size
}

func (q *queue) matchmaking() iter.Seq[[2]string] {
	n := q.nodes.root

	return func(yield func([2]string) bool) {
		q._matchmaking(n, yield)
	}
}

func (q *queue) _matchmaking(n *node, yield func([2]string) bool) {
	if n == q.nodes.leaf {
		return
	}

	// Find possible matches.
	matches := [4]*node{n.left, n.right, q.nodes.leaf, q.nodes.leaf}
	if n.left != q.nodes.leaf {
		matches[2] = q.nodes.findMax(n.left)
	}
	if n.right != q.nodes.leaf {
		matches[3] = q.nodes.findMin(n.right)
	}

	// Find the match which has the lowest rating gap.
	var best *node
	bestGap := GapLimit
	for _, match := range matches {
		// Skip leaf nodes.
		if match == q.nodes.leaf {
			continue
		}

		gap := math.Abs(n.rating - match.rating)
		if gap < bestGap {
			bestGap = gap
			best = match
		}
	}

	// Check does the best gap exceeds the allowed rating gaq.
	if best != nil && bestGap <= n.maxGap && bestGap <= best.maxGap {
		if !yield([2]string{n.id, best.id}) {
			return
		}

		// Remove matched nodes from nodes.
		q.nodes.remove(n)
		q.nodes.remove(best)

		// Call function recursively.
		q._matchmaking(q.nodes.root, yield)
		return
	}

	// Call function recursively on left and right subnodess.
	if n.left != q.nodes.leaf {
		q._matchmaking(n.left, yield)
	}

	if n.right != q.nodes.leaf {
		q._matchmaking(n.right, yield)
	}
}

// expandRatingGaps expands the allowed rating gap of each player so that players
// with larger rating gaps can eventually be paired together.
func (q *queue) expandRatingGaps() {
	if q.size < 1 {
		return
	}
	q._expandRatingGaps(q.nodes.root)
}

func (q *queue) _expandRatingGaps(n *node) {
	if n.maxGap < GapLimit {
		n.maxGap += DefaultMaxGap
	}

	if n.left != q.nodes.leaf {
		q._expandRatingGaps(n.left)
	}

	if n.right != q.nodes.leaf {
		q._expandRatingGaps(n.right)
	}
}
