package ws

import (
	"encoding/json"
	"justchess/internal/event"
	"justchess/internal/game"
	"log"
	"strings"
	"time"

	"github.com/treepeck/chego"
)

const (
	// How many seconds will empty room live.
	emptyDeadline = 5
)

type room struct {
	game       game.Game
	clients    map[string]*client
	register   chan *client
	unregister chan string
	handle     chan event.Event
	ticker     *time.Ticker
	timeToLive int
}

func newRoom(g game.Game) room {
	return room{
		game:       g,
		clients:    make(map[string]*client, 2),
		register:   make(chan *client),
		unregister: make(chan string),
		handle:     make(chan event.Event),
		ticker:     time.NewTicker(time.Second),
		timeToLive: emptyDeadline,
	}
}

func (r room) listenEvents(id string, remove chan<- string) {
	defer func() { remove <- id }()

	for {
		select {
		case c := <-r.register:
			r.add(c)

		case clientId := <-r.unregister:
			r.remove(clientId)

		case e := <-r.handle:
			switch e.Kind {
			case event.Chat:
				r.chat(e)

			case event.Move:
				var index byte
				if err := json.Unmarshal(e.Payload, &index); err != nil {
					log.Printf("invalid msg from client: %s", err)
					continue
				}
				if p, ok := r.game.Play(e.SenderId, index); ok {
					r.broadcast(event.JSON(event.Move, p))

					// If game has been terminated, broadcast EndPayload.
					end := r.game.EndPayload()
					if end.Termination != chego.Unterminated {
						r.broadcast(event.JSON(event.End, r.game.EndPayload()))
					}
				}

			case event.Resign:
				if r.game.Resign(e.SenderId) {
					r.broadcast(event.JSON(event.End, r.game.EndPayload()))
				}

			default:
				g, ok := r.game.(*game.RatedGame)
				if !ok {
					continue
				}

				sender := r.clients[e.SenderId]
				if sender == nil {
					continue
				}

				switch e.Kind {
				case event.OfferDraw:
					if oppId := g.OfferDraw(e.SenderId); len(oppId) != 0 {
						r.broadcast(event.JSON(event.Chat, sender.player.Name+" offers draw"))
						if opp := r.clients[oppId]; opp != nil {
							opp.send <- event.JSON(event.OfferDraw, nil)
						}
					}
				case event.AcceptDraw:
					if g.AcceptDraw(e.SenderId) {
						r.broadcast(event.JSON(event.End, r.game.EndPayload()))
						r.broadcast(event.JSON(event.Chat, sender.player.Name+" accepts draw"))
					}
				case event.DeclineDraw:
					if g.DeclineDraw(e.SenderId) {
						r.broadcast(event.JSON(event.Chat, sender.player.Name+" declines draw"))
					}
				}
			}

		case <-r.ticker.C:
			r.timeTick()
			if r.timeToLive == 0 {
				// Destroy the empty room.
				r.game.Abandon()
				return
			}
		}
	}
}

// add adds client to the room.
//
// Client will not be added if one of the following is true:
//   - The number of clients has reached the [clientsThreshold];
//   - The client with the same ID is already connected to the room.
func (r room) add(c *client) {
	if len(r.clients) == clientsThreshold {
		c.send <- event.JSON(event.Error, msgTooMany)
		return
	}
	if _, connected := r.clients[c.player.Id]; connected {
		c.send <- event.JSON(event.Error, msgConflict)
		return
	}

	r.clients[c.player.Id] = c
	r.game.Join(c.player.Id)

	c.forward = r.handle
	c.unregister = r.unregister

	// Send current game state to clients.
	c.send <- event.JSON(event.Game, r.game.GamePayload())

	// Broadcast number of online players.
	r.broadcast(event.JSON(event.ClientsCounter, len(r.clients)))
}

func (r room) remove(clientId string) {
	if _, connected := r.clients[clientId]; connected {
		delete(r.clients, clientId)
		r.game.Leave(clientId)

		// Broadcast number of online players.
		r.broadcast(event.JSON(event.ClientsCounter, len(r.clients)))
	} else {
		log.Printf("client %s isn't connected", clientId)
	}
}

func (r room) chat(e event.Event) {
	name := r.clients[e.SenderId].player.Name

	var b strings.Builder
	// Append sender's name.
	b.WriteString(name)
	b.WriteString(": ")
	// Append message.
	b.WriteString(strings.TrimSpace(strings.ReplaceAll(string(e.Payload), "\"", " ")))

	e.Payload = json.RawMessage(b.String())
	r.broadcast(event.JSON(event.Chat, b.String()))
}

func (r room) timeTick() {
	r.game.TimeTick()

	if len(r.clients) == 0 {
		r.timeToLive--
	}
}

func (r room) broadcast(raw []byte) {
	for _, c := range r.clients {
		c.send <- raw
	}
}
