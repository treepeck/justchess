// TODO: find out how people normally implement matchmaking.
package matchmaking

import (
	"github.com/treepeck/justchess/internal/randgen"
	"github.com/treepeck/justchess/internal/transport"
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

// queue provides the implementation of the matchmaking algorithm.
type queue struct {
	id     string
	in     chan proto.Message
	out    chan []byte
	ticker *time.Ticker
	nodes  *redBlackTree
	// Number of players.
	size int
}

// SpawnQueues spawns 9 queues and notifies transport package about it.
func SpawnQueues(ipc transport.Ipc) {
	// TODO: This is temporary code. It is needed to route the player from selected time
	// control to the queue because the matchmaking algorithm itself is kinda bad and it
	// requires a separate queue for every game mode.
	for _, id := range [9]string{"_b5DNjhhatG5", "8HoMa0GJMgvJ", "dVXdHdrQ-SBB", "56_e8-q4wDys", "RYwIsvOZNTDG", "77G2ZTJww_QK", "qq824814u_zm", "G_2xmQDExksB", "KNnUpq3G8hu4"} {
		q := &queue{
			id:     id,
			in:     make(chan proto.Message, 32),
			out:    ipc.Write,
			ticker: time.NewTicker(interval),
			nodes:  newRedBlackTree(),
		}
		go q.Listen()
		ipc.Add <- transport.Topic{
			Id: q.id,
			In: q.in,
		}
	}
}

func (q *queue) Listen() {
	for {
		select {
		case m := <-q.in:
			switch m.Kind {
			case proto.KindJoin:
				q.insert(m.Id)
			case proto.KindLeave:
				q.remove(m.Id)
			default:
				log.Printf("queue %s recieved invalid message from %s\n", q.id, m.Id)
			}
		case <-q.ticker.C:
			hasMatches := false
			for ids := range q.matchmaking() {
				hasMatches = true
				q.createMatch(ids)
			}
			if !hasMatches { // Do not send counter if no matches were found.
				continue
			}
		}

		log.Printf("%d\n", q.size)

		m := proto.Message{
			Id:      q.id,
			Kind:    proto.KindCounter,
			Payload: []byte{byte(q.size)}, // TODO: byte might not be enough.
		}
		log.Printf("kind: %d, id: %s, payload: %b\n", m.Kind, m.Id, m.Payload)
		q.out <- proto.Encode(m)
	}
}

// insert inserts a new player to the [queue] and returns it's size.
func (q *queue) insert(id string) int {
	rating := 1500.0 // TODO: get actual rating from database and cache it.

	log.Printf("player %s joins queue %s\n", id, q.id)

	q.nodes.insert(q.nodes.spawn(rating, id))
	q.size++
	return q.size
}

// remove removes an existing player from the [queue] and returns it's size.
func (q *queue) remove(id string) int {
	rating := 1500.0 // TODO: get actual rating from memory cache.
	n := search(q.nodes.root, rating, id)
	if n == nil {
		log.Printf("matchmaking: trying to remove non-existing player %s\n", id)
		return q.size
	}
	q.nodes.remove(n)
	q.size--
	log.Printf("player %s leaves queue %s\n", id, q.id)
	return q.size
}

func (q *queue) createMatch(playerIDs [2]string) {
	var payload [randgen.IdLen * 3]byte
	copy(payload[:randgen.IdLen], playerIDs[0])
	copy(payload[randgen.IdLen:2*randgen.IdLen], playerIDs[1])
	copy(payload[2*randgen.IdLen:], randgen.GenId(randgen.SeedLen)) // Game id.

	q.out <- proto.Encode(proto.Message{
		Id:      q.id,
		Kind:    proto.KindMatch,
		Payload: payload[:],
	})
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
