package ws

import (
	"log"
	"time"

	"github.com/treepeck/chego"
)

type room struct {
	// Connected clients.
	subscribers []string
	whiteId     string
	blackId     string
	ticker      *time.Ticker
	game        *chego.Game
}

func newRoom() *room {
	return &room{
		subscribers: make([]string, 0, 2),
		ticker:      time.NewTicker(time.Second),
		game:        chego.NewGame(),
	}
}

func (r *room) handle(ch <-chan clientEvent) {
	for {
		select {
		case e := <-ch:
			switch e.Action {
			case actionJoin:
				r.handlePlayerJoin(string(e.Payload))

			case actionLeave:
				r.handlePlayerLeave(string(e.Payload))

			case actionMakeMove:
				r.handleMakeMove()

			default:
				log.Print("invalid client event recieved")
			}

		case <-r.ticker.C:
			r.handleTimeTick()
		}
	}
}

func (r *room) handlePlayerJoin(id string) {
	r.subscribers = append(r.subscribers, id)
}

func (r *room) handlePlayerLeave(id string) {

}

func (r *room) handleMakeMove() {

}

func (r *room) handleTimeTick() {

}
