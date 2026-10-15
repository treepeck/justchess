// TODO: add support for different time controls which will require
// a full (probably) rewrite of this logic. But at the point at which the
// playerbase is currently we don't event need to have different time control
// because it will make it even more difficult for a few players to be paired
// together. For a time control selection the slider is ideal case because
// it has a 'default' value that most players will leave as-it and this means
// that players are more likely to be paired in a smaller playerbases.
// TODO: find out how people normally implement matchmaking.
package matchmaking

import (
	"github.com/treepeck/justchess/internal/randgen"
	"github.com/treepeck/justchess/internal/transport"
	"github.com/treepeck/justchess/pkg/db"
	"github.com/treepeck/justchess/pkg/proto"
	"iter"
	"log"
	"math/rand/v2"
	"time"
)

const (
	DefaultMaxGap float64 = 500
	GapLimit      float64 = 3000
	interval              = 3 * time.Second
)

const QueueId = "matchmakingq" // TODO: temporaty code.

// Queue provides the implementation of the matchmaking algorithm.
type Queue struct {
	gameRepo   db.GameRepo
	playerRepo db.PlayerRepo
	in         chan proto.Message
	ipc        transport.Ipc
	ticker     *time.Ticker
	nodes      *redBlackTree
	// Number of players.
	size uint16
}

func InitQueue(gr db.GameRepo, pr db.PlayerRepo, ipc transport.Ipc) *Queue {
	q := &Queue{
		gameRepo:   gr,
		playerRepo: pr,
		in:         make(chan proto.Message, 100),
		ipc:        ipc,
		ticker:     time.NewTicker(interval),
		nodes:      newRedBlackTree(),
	}

	// Add queue topic to the transport service.
	ipc.Add <- transport.Topic{
		Id: QueueId,
		In: q.in,
	}

	go q.listen()

	return q
}

func (q *Queue) listen() {
	for {
		select {
		case m := <-q.in:
			switch m.Kind {
			case proto.KindJoin:
				q.insert(string(m.Payload))
			case proto.KindLeave:
				q.remove(string(m.Payload))
			}
		case <-q.ticker.C:
			hasMatches := false
			for ids := range q.matchmaking() {
				hasMatches = true
				// Randomly select player's sides.
				whiteId := ids[0]
				blackId := ids[1]
				if rand.IntN(2) == 0 {
					whiteId = ids[1]
					blackId = ids[0]
				}
				q.createMatch(whiteId, blackId)
			}
			if !hasMatches {
				// Do not send counter if no matches were found.
				continue
			}
		}

		m := proto.Message{
			Id:      QueueId,
			Kind:    proto.KindCounter,
			Payload: proto.EncodeInt(q.size),
		}
		q.ipc.Write <- proto.Encode(m)
	}
}

// insert inserts a new player to the [Queue].
func (q *Queue) insert(id string) {
	rating := 1500.0 // TODO: get actual rating from database and cache it.

	log.Printf("player %s joins Queue %s\n", id, QueueId)

	q.nodes.insert(q.nodes.spawn(rating, id))
	log.Printf("%v\n", q.nodes)
	q.size++
}

// remove removes an existing player from the [Queue].
func (q *Queue) remove(id string) {
	rating := 1500.0 // TODO: get actual rating from memory cache.
	n := search(q.nodes.root, rating, id)
	if n == nil {
		log.Printf("matchmaking: trying to remove non-existing player %s\n", id)
		return
	}
	q.nodes.remove(n)
	q.size--
	log.Printf("player %s leaves Queue %s\n", id, QueueId)
}

func (q *Queue) createMatch(whiteId, blackId string) {
	gameId := randgen.GenId(randgen.SeedLen)

	// Encode the message payload.
	// First ID is white player ID.
	// Second ID is black player ID.
	// Third ID is the id of newly created game.
	var payload [3 * randgen.IdLen]byte
	copy(payload[:randgen.IdLen], whiteId)
	copy(payload[randgen.IdLen:2*randgen.IdLen], blackId)

	err := q.gameRepo.Insert(gameId, whiteId, blackId)
	if err != nil {
		q.ipc.Write <- proto.Encode(proto.Message{
			Id:      QueueId,
			Kind:    proto.KindError,
			Payload: payload[:2*randgen.IdLen],
		})
		return
	}
	// Add game id to payload.
	copy(payload[2*randgen.IdLen:], gameId)
	q.ipc.Write <- proto.Encode(proto.Message{
		Id:      QueueId,
		Kind:    proto.KindMatch,
		Payload: payload[:],
	})
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

		gap := n.rating - match.rating
		if gap < 0 {
			gap *= -1
		}
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
