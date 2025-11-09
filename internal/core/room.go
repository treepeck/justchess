package core

import (
	"log"
	"math/rand/v2"
	"time"

	"github.com/treepeck/chego"
	"github.com/treepeck/gatekeeper/pkg/event"
)

/*
roomState represents a domain of possible room states.
*/
type roomState int

const (
	// stateEmpty means that no players are connected.
	stateEmpty roomState = iota
	// stateWhite means that only white player is connected.
	stateWhite
	// stateBlack means that only black player is connected.
	stateBlack
	// stateBoth means that both players are connected.
	stateBoth

	// Empty room will live 20 seconds before destruction.
	emptyDeadline int = 20
)

/*
room wraps a single active game and additional info such as player ids to
act as a middleman between the [chego.Game] and connected players.
*/
type room struct {
	moves   []completedMove // Completed moves.
	id      string
	whiteId string
	blackId string
	move    chan moveReq
	join    chan string
	leave   chan string
	// Response channel to communicate with core.
	response chan<- []byte
	destroy  chan<- string
	// When timeToLive is equal to 0, the rool will send to destroy channel.
	timeToLive int
	// Number of active viewers.
	viewers int
	state   roomState
	clock   *time.Ticker
	game    *chego.Game
}

func newRoom(id, player1, player2 string, response chan<- []byte,
	destroy chan<- string, timeControl, timeBonus int) *room {
	r := &room{
		moves:      make([]completedMove, 0),
		id:         id,
		clock:      time.NewTicker(time.Second),
		timeToLive: emptyDeadline,
		move:       make(chan moveReq),
		join:       make(chan string),
		leave:      make(chan string),
		response:   response,
		destroy:    destroy,
		state:      stateEmpty,
		game:       chego.NewGame(),
	}

	// Randomly select players' sides.
	if rand.IntN(2) == 1 {
		r.whiteId = player1
		r.blackId = player2
	} else {
		r.whiteId = player2
		r.blackId = player1
	}

	r.game.WhiteTime = timeControl
	r.game.BlackTime = timeControl
	r.game.TimeBonus = timeBonus

	return r
}

func (r *room) handleEvents() {
	defer func() {
		r.clock.Stop()

		// Notify core server that the room was destroyed.
		r.destroy <- r.id
	}()

	for {
		select {
		case req := <-r.move:
			r.handleMove(req.playerId, req.move)

			if r.game.IsCheckmate() {
				log.Printf("checkmate in room %s", r.id)
				return
			}

		case playerId := <-r.join:
			r.handlePlayerJoin(playerId)

		case playerId := <-r.leave:
			r.handlePlayerLeave(playerId)

		case <-r.clock.C:
			r.handleTimeTick()

			if r.timeToLive == 0 {
				return
			}

			// if r.game.WhiteTime == 0 || r.game.BlackTime == 0 {
			// 	log.Printf("timeout in room %s", r.id)
			// 	return
			// }
		}
	}
}

func (r *room) handleMove(playerId string, m chego.Move) {
	// Deny the move if one of the following is true:
	//  - the game has been already over.
	//  - the player who send the move doesn't have rights to perform it.
	//  - the move isn't legal.
	if (r.game.Result != chego.ResultUnknown) ||
		(len(r.moves)%2 == 0 && playerId != r.whiteId) ||
		(len(r.moves)%2 != 0 && playerId != r.blackId) ||
		(!r.game.IsMoveLegal(m)) {
		return
	}

	r.moves = append(r.moves, completedMove{
		San:  r.game.PushMove(m),
		Move: m,
	})
}

func (r *room) handleTimeTick() {
	switch r.state {
	case stateEmpty:
		r.timeToLive--

	default:
		if len(r.moves)%2 == 0 {
			r.game.WhiteTime--
		} else {
			r.game.BlackTime--
		}
	}
}

func (r *room) handlePlayerJoin(playerId string) {
	switch playerId {
	case r.whiteId:
		r.state++

	case r.blackId:
		r.state += 2

	default:
		r.viewers++
	}

	log.Printf("player %s joined room %s", playerId, r.id)

	// Publish updated room info after player connection.
	raw, err := event.EncodeInternal(event.ActionRoomInfo, roomInfo{
		WhiteId: r.whiteId, BlackId: r.blackId, Viewers: r.viewers,
		TimeToLive: r.timeToLive,
	}, "", r.id)
	if err != nil {
		log.Printf("cannot encode internal event: %s", err)
		return
	}

	r.response <- raw
}

func (r *room) handlePlayerLeave(playerId string) {
	switch playerId {
	case r.whiteId:
		r.state--

	case r.blackId:
		r.state -= 2

	default:
		r.viewers--
	}

	log.Printf("player %s leaved room %s", playerId, r.id)

	// Publish updated room info after player disconnection.
	raw, err := event.EncodeInternal(event.ActionRoomInfo, roomInfo{
		WhiteId: r.whiteId, BlackId: r.blackId, Viewers: r.viewers,
		TimeToLive: r.timeToLive,
	}, "", r.id)
	if err != nil {
		log.Printf("cannot encode internal event: %s", err)
		return
	}

	r.response <- raw
}
