package core

import (
	"log"
)

/*
matchmaking implements the matchmaking system that finds a match if both players
have selected the similar time control and bonus.
Stores player id as a key and selected game paremeters as value.
TODO: don't pair two players if they have a large MMR gap.
*/
type matchmaking map[string]matchmakingDTO

func (mm matchmaking) enter(playerId string, entry matchmakingDTO) {
	mm[playerId] = entry
	log.Printf("player \"%s\" entered matchmaking", playerId)
}

func (mm matchmaking) leave(playerId string) {
	delete(mm, playerId)
	log.Printf("player \"%s\" leaved matchmaking", playerId)
}

func (mm matchmaking) hasEntered(playerId string) bool {
	_, ok := mm[playerId]
	return ok
}

/*
match returns player id if the player that has selected the similar game
parameters exists and an empty string otherwise.
*/
func (mm matchmaking) match(entry matchmakingDTO) string {
	for playerId, e := range mm {
		if e.TimeBonus == entry.TimeBonus || e.TimeControl == entry.TimeControl {
			log.Printf("match found \"%s\"", playerId)
			return playerId
		}
	}
	return ""
}
