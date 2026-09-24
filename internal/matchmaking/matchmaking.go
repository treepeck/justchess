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
	"github.com/treepeck/justchess/internal/transport"
	"github.com/treepeck/justchess/pkg/db"
	"github.com/treepeck/justchess/pkg/proto"
	// "math/rand/v2"
	"log"
	"time"
)

const interval = 5 * time.Second

type Service struct {
	queues      [9]*queue
	ipc         transport.Ipc
	playerRepo  db.PlayerRepo
	ticker      *time.Ticker
	ratingCache map[string]float64
}

func InitService(pr db.PlayerRepo, ipc transport.Ipc) Service {
	// Initialize the queues.
	var queues [9]*queue
	for i := range 9 {
		queues[i] = newQueue()
	}

	s := Service{
		playerRepo: pr,
		ipc:        ipc,
		queues:     queues,
		ticker:     time.NewTicker(interval),
	}
	go s.listen()
	return s
}

func (s Service) listen() {
	for {
		select {
		case m := <-s.ipc.ReadQueue:
			switch m.Payload.(type) {
			case proto.Join:
				s.register(m.PlayerId, m.Payload.(string))
			case proto.Leave:
				s.unregister(m.PlayerId, m.Payload.(string))
			}
		case <-s.ticker.C:
			for i, q := range s.queues {
				// TODO: do not overuse the plain string concatenation as it degrades the performance.
				url := "/queue/" + string(i+'0')
				for ids := range q.matchmaking() {
					s.onMatch(ids, url)
				}
				q.expandRatingGaps()
			}
		}
	}
}

// register inserts player into named queue.
// url should come in such format: '/queue/{id}'
func (s Service) register(playerId, url string) {
	ind := int(url[len(url)-1] - '0')
	if ind >= len(s.queues) {
		return
	}

	p, err := s.playerRepo.SelectProfile(playerId)
	if err != nil {
		log.Printf("player cannot be found %s: %v\n", err)
		return
	}

	s.ratingCache[playerId] = p.Rating

	s.queues[ind].insert(p.Rating, playerId)
	// Broadcast current players counter.
	s.ipc.Write <- proto.OutMessage{
		Recievers: []string{url}, // Pass url so that coordinator can broadcast to all connected clients.
		Payload:   proto.Counter(s.queues[ind].size),
	}
}

// unregister removes player from named queue.
func (s Service) unregister(playerId, url string) {
	ind := int(url[len(url)-1] - '0')
	if ind >= len(s.queues) {
		return
	}

	rating, ok := s.ratingCache[playerId]
	if !ok {
		log.Printf("rating of the player %s cannot be found in cache\n", playerId)
		return
	}

	s.queues[ind].remove(rating, playerId)
	// Broadcast current players counter.
	s.ipc.Write <- proto.OutMessage{
		Recievers: []string{url}, // Pass url so that coordinator can broadcast to all connected clients.
		Payload:   proto.Counter(s.queues[ind].size),
	}
}

func (s Service) onMatch(ids [2]string, url string) {
	// Randomly select players' sides.
	/*
		whiteId, blackId := ids[0], ids[1]
		if rand.IntN(2) == 1 {
			whiteId = ids[1]
			blackId = ids[0]
		}
	*/

	// TODO: create a game in game service.
	// TODO: send back redirect to the players.
	s.ipc.Write <- proto.OutMessage{
		Recievers: ids[:],
		Payload:   proto.Redirect(url),
	}
}
