package core

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"time"

	"github.com/treepeck/chego"
	"github.com/treepeck/gatekeeper/pkg/types"
)

/*
room stores and manages a single active game state and additional info such as
player ids and game parameters.
*/
type room struct {
	id, whiteId, blackId   string
	state                  roomState
	move                   chan moveDTO
	forward                chan<- types.MetaEvent
	timeControl, timeBonus int
	// Will tick every second and decrement the destroyTime when all players are
	// disconnected.
	ticker      *time.Ticker
	destroyTime int
	game        *chego.Game
}

/*
newRoom randomly creates a new room, selects the player sides, and starts the
game clock.
*/
func newRoom(
	id, player1Id, player2Id string,
	forward chan<- types.MetaEvent,
	timeControl, timeBonus int,
) *room {
	r := &room{
		id:          id,
		ticker:      time.NewTicker(time.Second),
		destroyTime: 20,
		move:        make(chan moveDTO),
		forward:     forward,
		timeControl: timeControl,
		timeBonus:   timeBonus,
		state:       stateEmpty,
	}

	// Randomly select players' colors.
	if rand.IntN(2) == 1 {
		r.whiteId = player1Id
		r.blackId = player2Id
	} else {
		r.whiteId = player2Id
		r.blackId = player1Id
	}

	r.game = chego.NewGame()

	// TimeControl equal to 0 means that the game does not have time
	// restrictions.
	if r.timeControl > 0 {
		r.game.SetClock(r.timeControl, r.timeBonus)
	}

	return r
}

/*
run is needed to ensure that concurrent events are handled sequentially.
This prevents race conditions: ror example, when the time ticks, the room does
not handles the move.
*/
func (r *room) run(roomId string) {
	defer func() {
		r.ticker.Stop()
		r.game.Clock.Stop()
	}()

	for {
		select {
		case dto := <-r.move:
			r.handleMove(roomId, dto)

		case <-r.ticker.C:
			r.destroyTime--
			if r.destroyTime <= 0 {
				log.Print("room destroyed")
				r.forward <- types.MetaEvent{
					Action:  types.ActionRemoveRoom,
					Payload: nil,
					RoomId:  roomId,
				}
				// TODO: notify the core that the room is destroyed.
				return
			}

		case <-r.game.Clock.C:
			if r.game.Position.ActiveColor == chego.ColorWhite {
				r.game.WhiteTime--

				if r.game.WhiteTime <= 0 {
					// TODO: end game.
				}
			} else {
				r.game.BlackTime--

				if r.game.BlackTime <= 0 {
					// TODO: end game.
				}
			}
		}
	}
}

func (r *room) handleJoin(playerId string) {
	switch r.state {
	case stateEmpty:
		switch playerId {
		case r.whiteId:
			r.state = stateWhite
		case r.blackId:
			r.state = stateBlack
		}
	case stateWhite:
		if playerId == r.blackId {
			r.state = stateBoth

			r.ticker.Stop()
		}
	case stateBlack:
		if playerId == r.whiteId {
			r.state = stateBoth

			r.ticker.Stop()
		}
	}
}

func (r *room) handleLeave(playerId string) {
	switch r.state {
	case stateBoth:
		switch playerId {
		case r.whiteId:
			r.state = stateBlack

			r.ticker.Reset(time.Second)
		case r.blackId:
			r.state = stateWhite

			r.ticker.Reset(time.Second)
		}
	case stateWhite:
		if playerId == r.whiteId {
			r.state = stateEmpty
		}

	case stateBlack:
		if playerId == r.blackId {
			r.state = stateEmpty
		}
	}
}

func (r *room) handleMove(roomId string, dto moveDTO) {
	// Deny the move if one of the following is true:
	//  * the game has been already over.
	//  * the player who send the move doesn't have rights to perform it.
	//  * the move isn't legal.
	if (r.game.Result != chego.ResultUnscored) ||
		(len(r.game.MoveStack)%2 == 0 && dto.playerId != r.whiteId) ||
		(len(r.game.MoveStack)%2 != 0 && dto.playerId != r.blackId) ||
		(!r.game.IsMoveLegal(dto.move)) {
		return
	}

	r.game.PushMove(dto.move)

	cm := r.game.MoveStack[len(r.game.MoveStack)-1]

	// Compress legal moves into a smaller slice to reduce the event size.
	lm := make([]int, r.game.LegalMoves.LastMoveIndex)
	for i := range r.game.LegalMoves.LastMoveIndex {
		lm[i] = int(r.game.LegalMoves.Moves[i])
	}

	p, err := json.Marshal(types.CompletedMove{
		LegalMoves: lm,
		San:        cm.San,
		Fen:        cm.Fen,
		TimeLeft:   cm.TimeLeft,
	})
	if err != nil {
		log.Printf("cannot encode completed move payload: %s", err)
		return
	}

	r.forward <- types.MetaEvent{
		Action:  types.ActionCompletedMove,
		Payload: p,
		RoomId:  roomId,
	}
}
