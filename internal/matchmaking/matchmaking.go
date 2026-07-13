// Package matchmaking implements the matchmaking algorithm. Its purpose is to
// find the best pairs of players in a pool containing all players who have
// selected the same time control. The best pair is defined as the pair with
// the smallest possible MMR gap. The algorithm runs periodically at the
// interval defined by the [Interval] const.
//
// Each player in the matchmaking pool has their own allowed MMR gap. This
// prevents players with large MMR gaps from being paired when the pool
// contains only a small number of players. Each player's allowed MMR gap
// increases every [matchmakingInterval]. That is, the longer a player waits
// in the queue, the larger their allowed MMR gap becomes. As a result, players
// with large MMR differences will eventually be paired when the pool is nearly empty.
//
// Currently, only MMR is taken into account. The algorithm can be extended to
// also consider network latency, rating deviation, and volatility.
package matchmaking

import (
	"fmt"
	"iter"
	"math"
	"time"
)

const (
	Interval              = 5 * time.Second
	DefaultMaxGap float64 = 500
	GapLimit      float64 = 3000
)

// Profile is a player's matchmaking profile.
type Profile struct {
	Id  string
	MMR float64
}

// Pool provides the implementation of the matchmaking algorithm.
// WARN: it's the caller's responsibility to ensure thread-safetiness.
type Pool struct {
	nodes  *redBlackTree
	ticker *time.Ticker
	// Number of players.
	size int
}

func NewPool() *Pool {
	return &Pool{
		nodes:  newRedBlackTree(),
		ticker: time.NewTicker(Interval),
	}
}

// Insert inserts a new player to the [Pool] and returns it's size.
func (p *Pool) Insert(mmr float64, id string) int {
	p.nodes.insert(p.nodes.spawn(mmr, id))
	p.size++
	return p.size
}

// remove removes an existing player from the [Pool] and returns it's size.
func (p *Pool) Remove(mmr float64, id string) int {
	n := search(p.nodes.root, mmr, id)
	if n == nil {
		fmt.Printf("matchmaking: trying to remove non-existing player \"%s\"\n", id)
		return p.size
	}
	p.nodes.remove(n)
	p.size--
	return p.size
}

func (p *Pool) Matchmaking() iter.Seq[[2]string] {
	n := p.nodes.root

	return func(yield func([2]string) bool) {
		p.matchmaking(n, yield)
	}
}

func (p *Pool) matchmaking(n *node, yield func([2]string) bool) {
	if n == p.nodes.leaf {
		return
	}

	// Find possible matches.
	matches := [4]*node{n.left, n.right, p.nodes.leaf, p.nodes.leaf}
	if n.left != p.nodes.leaf {
		matches[2] = p.nodes.findMax(n.left)
	}
	if n.right != p.nodes.leaf {
		matches[3] = p.nodes.findMin(n.right)
	}

	// Find the match which has the lowest rating gap.
	var best *node
	bestGap := GapLimit
	for _, match := range matches {
		// Skip leaf nodes.
		if match == p.nodes.leaf {
			continue
		}

		gap := math.Abs(n.mmr - match.mmr)
		if gap < bestGap {
			bestGap = gap
			best = match
		}
	}

	// Check does the best gap exceeds the allowed rating gap.
	if best != nil && bestGap <= n.maxGap && bestGap <= best.maxGap {
		if !yield([2]string{n.id, best.id}) {
			return
		}

		// Remove matched nodes from nodes.
		p.nodes.remove(n)
		p.nodes.remove(best)

		// Call function recursively.
		p.matchmaking(p.nodes.root, yield)
		return
	}

	// Call function recursively on left and right subnodess.
	if n.left != p.nodes.leaf {
		p.matchmaking(n.left, yield)
	}

	if n.right != p.nodes.leaf {
		p.matchmaking(n.right, yield)
	}
}

// ExpandRatingGaps expands the allowed MMR gap of each player so that players
// with larger rating gaps can eventually be paired together.
func (p *Pool) ExpandMMRGaps() {
	if p.size < 1 {
		return
	}
	p.expandMMRGaps(p.nodes.root)
}

func (p *Pool) expandMMRGaps(n *node) {
	if n.maxGap < GapLimit {
		n.maxGap += DefaultMaxGap
	}

	if n.left != p.nodes.leaf {
		p.expandMMRGaps(n.left)
	}

	if n.right != p.nodes.leaf {
		p.expandMMRGaps(n.right)
	}
}
