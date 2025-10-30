package core

import (
	"log"
	"math/rand/v2"
	"strconv"
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
	deadline int = 120
)

/*
room stores and manages a single active game state and additional info such as
player ids and game parameters.
*/
type room struct {
	id, whiteId, blackId string
	state                roomState
	move                 chan moveDTO
	forward              chan<- event.Internal
	destroyTime          int
	// clock will tick every second and decrement the destroyTime when all
	// players are disconnected.
	clock       *time.Ticker
	timeControl int
	game        *chego.Game
}

/*
newRoom creates a new room, randomly selects the player sides, and starts the
clock.
*/
func newRoom(
	id, player1Id, player2Id string,
	forward chan<- event.Internal,
	timeControl, timeBonus int,
) *room {
	r := &room{
		id:          id,
		clock:       time.NewTicker(time.Second),
		destroyTime: deadline,
		move:        make(chan moveDTO),
		forward:     forward,
		timeControl: timeControl,
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

	return r
}

/*
run is needed to ensure that concurrent events are handled sequentially.
This prevents race conditions: for example, when the time ticks, the room doesn't
handles the move.
*/
func (r *room) run() {
	defer func() {
		// Notify the core server that the room was destroyed.
		r.forward <- event.Internal{
			Action:  event.ActionRemoveRoom,
			Payload: []byte(strconv.Quote(r.id)),
		}

		r.clock.Stop()
	}()

	for {
		select {
		case dto := <-r.move:
			r.handleMove(dto)
			if r.game.IsCheckmate() {
				log.Print("checkmate")
				return
			}

		case <-r.clock.C:
			r.handleTimeTick()
		}
	}
}

func (r *room) handleMove(dto moveDTO) {
	// Deny the move if one of the following is true:
	//  - the game has been already over.
	//  - the player who send the move doesn't have rights to perform it.
	//  - the move isn't legal.
	if (r.game.Result != chego.ResultUnknown) ||
		(len(r.game.CompletedMoves)%2 == 0 && dto.playerId != r.whiteId) ||
		(len(r.game.CompletedMoves)%2 != 0 && dto.playerId != r.blackId) ||
		(!r.game.IsMoveLegal(dto.move)) {
		return
	}

	r.game.PushMove(dto.move)

	// cm := r.game.CompletedMoves[len(r.game.CompletedMoves)-1]

	// // Compress legal moves into a smaller slice to reduce the message size.
	// lm := make([]chego.Move, r.game.LegalMoves.LastMoveIndex)
	// for i := range r.game.LegalMoves.LastMoveIndex {
	// 	lm[i] = r.game.LegalMoves.Moves[i]
	// }

	// p, err := json.Marshal(types.CompletedMove{
	// 	LegalMoves: lm,
	// 	San:        cm.San,
	// 	Fen:        cm.Fen,
	// 	TimeLeft:   cm.TimeLeft,
	// })
	// if err != nil {
	// 	log.Printf("cannot encode completed move payload: %s", err)
	// 	return
	// }

	// r.forward <- types.ServerEvent{
	// 	Action:  types.ActionCompletedMove,
	// 	Payload: p,
	// 	RoomId:  roomId,
	// }
}

func (r *room) handleTimeTick() {
	switch r.state {

	}
}
