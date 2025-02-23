package ws

import (
	"justchess/pkg/game"
	"justchess/pkg/game/enums"
	"log"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// upgrader is used to upgrate the HTTP connection into the websocket protocol.
var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	// clientEvents is used to synchronize and handle concurrent messages (client`s registration and unregistration).
	clientEvents chan clientEvent
	// All connected clients.
	clients map[uuid.UUID]*client
	// gameEvents is used to synchronize and handle concurrent messages (game creation and delition).
	gameEvents chan gameEvent
	// All active games.
	games map[uuid.UUID]*game.Game
}

func NewHub() *Hub {
	return &Hub{
		clientEvents: make(chan clientEvent),
		clients:      make(map[uuid.UUID]*client),
		gameEvents:   make(chan gameEvent),
		games:        make(map[uuid.UUID]*game.Game),
	}
}

// HandleNewConnection creates a new client and registers it in the Hub.
func (h *Hub) HandleNewConnection(rw http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	id := uuid.New()
	c := newClient(conn, h)

	go c.readPump(id)
	go c.writePump(id)

	h.clientEvents <- clientEvent{id: id, sender: c, eType: REGISTER}
}

// EventPump pumps incomming events from the channels and handles them.
func (h *Hub) EventPump() {
	for {
		select {
		case e := <-h.clientEvents:
			h.handleClientEvent(e)

		case e := <-h.gameEvents:
			h.handleGameEvent(e)
		}
	}
}

func (h *Hub) handleClientEvent(e clientEvent) {
	switch e.eType {
	case REGISTER:
		h.addClient(e.id, e.sender)
		h.broadcastClientsCounter()

	case UNREGISTER:
		h.removeClient(e.id)
		h.broadcastClientsCounter()
	}
}

func (h *Hub) handleGameEvent(e gameEvent) {
	sender, _ := uuid.FromBytes(e.payload[:16])
	switch e.eType {
	case GET_AVAILIBLE:
		h.sendAvailibleGames(sender)

	case CREATE:
		h.addGame(sender, e.payload[16], e.payload[17])

	case JOIN:
		gameId, _ := uuid.FromBytes(e.payload[16:32])
		h.handleJoinGame(sender, gameId)

	case GET:
		h.handleGetGame(sender)

	}
}

func (h *Hub) addClient(id uuid.UUID, c *client) {
	h.clients[id] = c
	log.Printf("client %s added\n", id.String())
}

// removeClient removes the client from the hub and closes it`s channel.
func (h *Hub) removeClient(id uuid.UUID) {
	if c, ok := h.clients[id]; ok {
		close(c.send)
		delete(h.clients, id)
		log.Printf("client %s removed\n", id.String())
	}
}

func (h *Hub) addGame(sender uuid.UUID, control, bonus byte) {
	id := uuid.New()
	g := game.NewGame(nil, control, bonus)

	if rand.Intn(2) == 1 {
		g.WhiteId = sender
	} else {
		g.BlackId = sender
	}

	h.games[id] = g
	log.Printf("game %s added\n", id.String())

	h.broadcastAddGame(id, g)
}

func (h *Hub) handleJoinGame(sender, gameId uuid.UUID) {
	if g, ok := h.games[gameId]; ok {
		if g.Status == enums.NotStarted {
			if g.WhiteId == uuid.Nil {
				g.WhiteId = sender
			} else {
				g.BlackId = sender
			}
			g.Status = enums.Continues

			// Redirect the players to the play page.
			h.sendRedirect(g.WhiteId, gameId)
			h.sendRedirect(g.BlackId, gameId)
			log.Printf("whiteId: %s, blackId: %s\n", g.WhiteId.String(), g.BlackId.String())
		}
		return
	}
}

func (h *Hub) handleGetGame(id uuid.UUID) {
	for _, g := range h.games {
		if g.WhiteId == id || g.BlackId == id {
			h.sendGameInfo(id, g)
			return
		}
	}
}

func (h *Hub) removeGame(id uuid.UUID) {
	if _, ok := h.games[id]; ok {
		delete(h.games, id)
		log.Printf("game %s removed\n", id.String())
		h.broadcastRemoveGame(id)
		return
	}
}

func (h *Hub) broadcastClientsCounter() {
	// To send larger numbers, such as uint32, use 5 bytes.
	msg := make([]byte, 5)
	l := len(h.clients)
	msg[0] = uint8(l) & 0xF
	msg[1] = uint8(l>>8) & 0xF
	msg[2] = uint8(l>>16) & 0xF
	msg[3] = uint8(l>>24) & 0xF
	msg[4] = CLIENTS_COUNTER
	for _, c := range h.clients {
		c.send <- msg
	}
}

func (h *Hub) broadcastAddGame(id uuid.UUID, g *game.Game) {
	msg := make([]byte, 19)
	copy(msg[:16], id[:])
	msg[16] = g.TimeControl
	msg[17] = g.TimeBonus
	msg[18] = ADD_GAME
	for _, c := range h.clients {
		c.send <- msg
	}
}

func (h *Hub) broadcastRemoveGame(id uuid.UUID) {
	msg := make([]byte, 17)
	copy(msg[:16], id[:])
	msg[16] = REMOVE_GAME
	for _, c := range h.clients {
		c.send <- msg
	}
}

func (h *Hub) sendRedirect(reciever, to uuid.UUID) {
	if c, ok := h.clients[reciever]; ok {
		msg := make([]byte, 17)
		copy(msg[:16], to[:])
		msg[16] = REDIRECT
		c.send <- msg
		return
	}
}

func (h *Hub) sendAvailibleGames(reciever uuid.UUID) {
	if c, ok := h.clients[reciever]; ok {
		cnt := 0
		for id, g := range h.games {
			if cnt == 10 {
				return
			}
			msg := make([]byte, 19)
			copy(msg[:16], id[:])
			msg[16] = g.TimeControl
			msg[17] = g.TimeBonus
			msg[18] = ADD_GAME
			c.send <- msg
			cnt++
		}
	}
}

func (h *Hub) sendGameInfo(reciever uuid.UUID, g *game.Game) {
	msg := make([]byte, 34)
	copy(msg[:16], g.WhiteId[:])
	copy(msg[16:32], g.BlackId[:])
	msg[32] = byte(g.Status)
	msg[33] = byte(g.Result)

	for i, move := range g.Moves {
		if i != 0 {
			msg = append(msg, 0xFF) // Separator.
		}
		msg = append(msg, []byte(move.SAN)...)
		msg = append(msg, 0xFF) // Separator.
		msg = append(msg, []byte(move.FEN)...)
	}
	msg = append(msg, GAME_INFO)

	if c, ok := h.clients[reciever]; ok {
		c.send <- msg
	}
}
