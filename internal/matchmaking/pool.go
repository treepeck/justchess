package matchmaking

import "log"

type TimeControl struct {
	Control int `json:"c"`
	Bonus   int `json:"b"`
}

var limits = []TimeControl{
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

type Ticket struct {
	Control TimeControl
	Rating  float64
}

type JoinDTO struct {
	PlayerId string
	Ticket   Ticket
}

type Pool struct {
	Join    chan JoinDTO
	Leave   chan string
	tickets map[string]Ticket
	forest  map[TimeControl]*redBlackTree
}

func NewPool() Pool {
	forest := make(map[TimeControl]*redBlackTree, 9)

	// Add trees.
	for _, limit := range limits {
		forest[limit] = newRedBlackTree()
	}

	return Pool{
		Join:    make(chan JoinDTO),
		Leave:   make(chan string),
		forest:  forest,
		tickets: make(map[string]Ticket),
	}
}

func (p Pool) EventBus() {
	for {
		select {
		case dto := <-p.Join:
			p.handleJoin(dto)

		case playerId := <-p.Leave:
			p.handleLeave(playerId)
		}
	}
}

func (p Pool) handleJoin(dto JoinDTO) {
	// Deny the request if the player already have opened a ticket.
	if _, exists := p.tickets[dto.PlayerId]; exists {
		log.Printf("Player %s tries to open multiple tickets", dto.PlayerId)
		return
	}

	tree, exists := p.forest[dto.Ticket.Control]
	if !exists {
		return
	}
	n := tree.spawn(dto.Ticket.Rating, dto.PlayerId)
	tree.insertNode(n)

	// Add matchmaking ticket.
	p.tickets[dto.PlayerId] = dto.Ticket

	log.Printf("Player %s joins matchmaking", dto.PlayerId)
}

func (p Pool) handleLeave(playerId string) {
	ticket, exists := p.tickets[playerId]
	if !exists {
		return
	}

	tree, exists := p.forest[ticket.Control]
	if !exists {
		return
	}

	n := tree.search(ticket.Rating, playerId)
	tree.removeNode(n)

	// Remove matchmaking ticket.
	delete(p.tickets, playerId)

	log.Printf("Player %s leaves matchmaking", playerId)
}
