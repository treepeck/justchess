package core

import (
	"encoding/json"
	"log"
	"net/http"

	"justchess/internal/auth"
	"justchess/internal/game"
	"justchess/internal/mm"

	"github.com/BelikovArtem/chego"
	"github.com/BelikovArtem/gatekeeper/pkg/event"
	"github.com/BelikovArtem/gatekeeper/pkg/mq"
	"github.com/rabbitmq/amqp091-go"
)

type Core struct {
	channel     *amqp091.Channel
	Bus         chan event.InternalEvent
	gameRooms   map[string]*game.GameRoom
	matchmaking *mm.Matchmaking
}

/*
NewCore opens a core channel, declares the "hub" exchange and creates a new Core
instance.
*/
func NewCore(ch *amqp091.Channel) *Core {
	c := &Core{
		channel:     ch,
		Bus:         make(chan event.InternalEvent),
		gameRooms:   make(map[string]*game.GameRoom),
		matchmaking: mm.NewMatchmaking(),
	}
	return c
}

/*
Handle consequentially (one at a time) accepts events from the Bus and forwards
them to the corresponding handler function.
*/
func (c *Core) Handle() {
	for {
		e := <-c.Bus
		switch e.Action {
		// case event.CREATE_ROOM:

		default:
			log.Print("unacceptable event: %v", e)
		}
	}
}

func (c *Core) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /mm", auth.AuthorizeRequest(c.handleMatchmaking))
	return mux
}

func (c *Core) handleMatchmaking(rw http.ResponseWriter, r *http.Request) {
	var p mm.WaitRoom
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(rw, "Malformed request body.", http.StatusBadRequest)
		return
	}

	c.Bus <- event.InternalEvent{}
}

/*
handleEvent handles the consumed client event from the gate queue.
*/
func (c *Core) handleEvent(e event.InternalEvent) error {
	return nil
}

func (c *Core) addGameRoom(creatorId string, p createRoomPayload) {
	gr := game.NewGameRoom(p.TimeControl, p.TimeBonus)

	c.gameRooms[gr.Id] = gr

	raw := event.EncodeOrPanic(event.InternalEvent{
		Action: event.ADD_ROOM,
		Payload: event.EncodeOrPanic(addRoomPayload{
			Id:          gr.Id,
			CreatorId:   creatorId,
			TimeControl: p.TimeControl,
			TimeBonus:   p.TimeBonus,
		}),
		RoomId: "hub",
	})
	mq.Publish(c.channel, "core", raw)
}

func (c *Core) handleMakeMove(m chego.Move) {

}
