// TODO: find out how people normally implement matchmaking.
package matchmaking

import (
	"github.com/treepeck/justchess/pkg/proto"
	"iter"
	"log"
	"math"
	"time"
)

const (
	DefaultMaxGap float64 = 500
	GapLimit      float64 = 3000
	interval              = 3 * time.Second
)

// Queue provides the implementation of the matchmaking algorithm.
type Queue struct {
	id     string
	In     chan proto.Message
	ticker *time.Ticker
	nodes  *redBlackTree
	// Number of players.
	size int
}

func SpawnQueue(id string) chan proto.Message {
	q := &Queue{
		id:     id,
		In:     make(chan proto.Message, 32),
		ticker: time.NewTicker(interval),
		nodes:  newRedBlackTree(),
	}

	go q.Listen()

	return q.In
}

func (q *Queue) Listen() {
	for {
		select {
		case m := <-q.In:
			switch m.Kind {
			case proto.KindJoin:
				q.insert(m.Id)
			case proto.KindLeave:
				q.remove(m.Id)
			default:
				log.Printf("queue %s recieved invalid message from %s\n", q.id, m.Id)
			}
		case <-q.ticker.C:
			q.matchmaking()
		}
	}
}

// insert inserts a new player to the [Queue] and returns it's size.
func (q *Queue) insert(id string) int {
	rating := 1500.0 // TODO: get actual rating from database and cache it.

	log.Printf("player %s joins queue %s\n", id, q.id)

	q.nodes.insert(q.nodes.spawn(rating, id))
	q.size++
	return q.size
}

// remove removes an existing player from the [Queue] and returns it's size.
func (q *Queue) remove(id string) int {
	rating := 1500.0 // TODO: get actual rating from memory cache.
	n := search(q.nodes.root, rating, id)
	if n == nil {
		log.Printf("matchmaking: trying to remove non-existing player \"%s\"\n", id)
		return q.size
	}
	q.nodes.remove(n)
	q.size--

	log.Printf("player %s leaves queue %s\n", id, q.id)
	return q.size
}

func (q *Queue) matchmaking() iter.Seq[[2]string] {
	n := q.nodes.root

	return func(yield func([2]string) bool) {
		q._matchmaking(n, yield)
	}
}

func (q *Queue) _matchmaking(n *node, yield func([2]string) bool) {
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

	// Check does the best gap exceeds the allowed rating gap.
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

	// Call function recursively on left and right subnodes.
	if n.left != q.nodes.leaf {
		q._matchmaking(n.left, yield)
	}

	if n.right != q.nodes.leaf {
		q._matchmaking(n.right, yield)
	}
}

// expandRatingGaps expands the allowed rating gap of each player so that players
// with larger rating gaps can eventually be paired together.
func (q *Queue) expandRatingGaps() {
	if q.size < 1 {
		return
	}
	q._expandRatingGaps(q.nodes.root)
}

func (q *Queue) _expandRatingGaps(n *node) {
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
