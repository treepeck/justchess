package matchmaking

import (
	"log"
)

// Ticket defines the matchmaking criterias. Only players with simmilar tickers
// can be matched together.
type Ticket struct {
	// Number of minutes each player has at the beginning of the game.
	Control int `json:"c"`
	// Number of seconds added after each completed move.
	Bonus int `json:"b"`
}

type Pool struct {
	tickets map[string]Ticket
	forest  map[Ticket]*redBlackTree
}

func NewPool() Pool {
	forest := make(map[Ticket]*redBlackTree, 9)

	// Predefine matchmaking tickets for allowed game modes.
	tickets := []Ticket{
		{Control: 1, Bonus: 0},
		{Control: 2, Bonus: 1},
		{Control: 3, Bonus: 0},
		{Control: 3, Bonus: 2},
		{Control: 5, Bonus: 0},
		{Control: 5, Bonus: 2},
		{Control: 10, Bonus: 0},
		{Control: 10, Bonus: 10},
		{Control: 15, Bonus: 10},
	}
	// Make a tree for every type of ticket.
	for _, t := range tickets {
		forest[t] = newRedBlackTree()
	}

	return Pool{
		forest:  forest,
		tickets: make(map[string]Ticket),
	}
}

func (p Pool) Join(id string, rating float64, t Ticket) {
	// Deny the request if the player already have opened a ticket.
	if _, exists := p.tickets[id]; exists {
		log.Printf("Player %s tries to open multiple tickets", id)
		return
	}

	tree, exists := p.forest[t]
	if !exists {
		return
	}
	n := tree.spawn(rating, id)
	tree.insertNode(n)

	// Add matchmaking ticket.
	p.tickets[id] = t

	log.Printf("Player %s joins matchmaking", id)
}

func (p Pool) Leave(id string, rating float64) {
	t, exists := p.tickets[id]
	if !exists {
		return
	}

	tree, exists := p.forest[t]
	if !exists {
		return
	}

	n := tree.search(rating, id)
	tree.removeNode(n)

	// Remove matchmaking ticket.
	delete(p.tickets, id)

	log.Printf("Player %s leaves matchmaking", id)
}
