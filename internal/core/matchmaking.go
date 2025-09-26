package core

import (
	"log"

	"github.com/treepeck/gatekeeper/pkg/types"
)

/*
matchmaking implements the matchmaking system that finds a match if both players
have selected the similar time control and bonus.
TODO: do not pair two players if they have a large MMR gap.
*/
type matchmaking struct {
	// pool stores player id as a key and selected game paremeters as value.
	pool map[string]types.EnterMatchmaking
}

func newMatchmaking() *matchmaking {
	return &matchmaking{pool: make(map[string]types.EnterMatchmaking)}
}

func (m *matchmaking) enter(playerId string, entry types.EnterMatchmaking) {
	m.pool[playerId] = entry
	log.Printf("player \"%s\" entered matchmaking", playerId)
}

func (m *matchmaking) leave(playerId string) {
	delete(m.pool, playerId)
	log.Printf("player \"%s\" leaved matchmaking", playerId)
}

func (m *matchmaking) hasEntered(playerId string) bool {
	_, ok := m.pool[playerId]
	return ok
}

/*
match returns player id if the player that has selected the similar game
parameters exists and an empty string otherwise.
*/
func (m *matchmaking) match(entry types.EnterMatchmaking) string {
	for playerId, e := range m.pool {
		if e.TimeBonus == entry.TimeBonus || e.TimeControl == entry.TimeControl {
			log.Printf("match found \"%s\"", playerId)
			return playerId
		}
	}
	return ""
}
