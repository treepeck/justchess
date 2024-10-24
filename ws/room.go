package ws

import (
	"chess-api/models/game"
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
	"chess-api/models/user"
	"encoding/json"
	"log/slog"
	"math/rand"

	"github.com/google/uuid"
)

// Room stores two players and a game they play.
// The Room type is very similar to the Manager.
// There is always one single Room for the every game.
type Room struct {
	Id         uuid.UUID     `json:"id"`
	Game       *game.G       `json:"game"`
	Owner      user.U        `json:"owner"`
	Bonus      uint          `json:"bonus"`
	Control    enums.Control `json:"control"`
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	close      chan bool // channel to close the room
}

// CreateRoomDTO provides necessary data to register a new Room.
type CreateRoomDTO struct {
	Control enums.Control `json:"control"`
	Bonus   uint          `json:"bonus"`
	Owner   user.U        `json:"owner"`
}

// newRoom creates and runs a new room. The owner client is added to the room.
func newRoom(cr CreateRoomDTO, owner *Client) *Room {
	r := &Room{
		Id:         uuid.New(),
		Game:       nil,
		Owner:      cr.Owner,
		Bonus:      cr.Bonus,
		Control:    cr.Control,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		close:      make(chan bool),
	}
	go r.run()
	r.addClient(owner)
	return r
}

func (r *Room) run() {
	for {
		select {
		case c := <-r.register:
			r.addClient(c)

		case c := <-r.unregister:
			r.removeClient(c)

		case <-r.close:
			// exit loop
			return
		}
	}
}

// addClient adds a client to the room.
func (r *Room) addClient(c *Client) {
	// if the game hasn`t started yet.
	if r.Game == nil {
		r.clients[c] = true
	} else {
		// handle reconnections ONLY and deny any other clients to connect.
		if r.Game.WhiteId == c.User.Id ||
			r.Game.BlackId == c.User.Id {
			r.clients[c] = true
		} else {
			return
		}
	}
	c.currentRoom = r
	c.sendRedirect(r.Id)

	r.startGame()
}

// removeClient deletes the client from the room and deletes the room itself if the
// room owner disconnects and the game has not been started yet (game aborted).
func (r *Room) removeClient(c *Client) {
	delete(r.clients, c)
	c.currentRoom = nil
	if (r.Game == nil && r.Owner.Id == c.User.Id) || len(r.clients) == 0 {
		c.manager.remove <- r
	}
}

// startGame creates a new game if all clients are connected.
func (r *Room) startGame() {
	if r.Game != nil || len(r.clients) != 2 {
		return
	}

	// randomize side selection
	var whiteId uuid.UUID
	var blackId uuid.UUID

	players := make([]*Client, 0)
	for c := range r.clients {
		players = append(players, c)
	}

	if rand.Intn(100) < 50 {
		whiteId = players[0].User.Id
		blackId = players[1].User.Id
	} else {
		whiteId = players[1].User.Id
		blackId = players[0].User.Id
	}

	r.Game = game.NewG(r.Id, r.Control, r.Bonus, whiteId, blackId)
	r.broadcast(UPDATE_GAME)
}

func (r *Room) broadcast(action string) {
	fn := slog.String("func", "room.broadcast")

	var payload []byte
	var err error

	switch action {
	case UPDATE_GAME:
		payload, err = json.Marshal(r.Game)

	default:
		slog.Warn("event had unknown action", fn, "action", action)
		return
	}

	if err != nil {
		slog.Warn("cannot Marshal payload", fn, "err", err)
	}
	e := Event{
		Action:  action,
		Payload: payload,
	}
	for c := range r.clients {
		c.writeEventBuffer <- e
	}

}

func (r *Room) handleTakeMove(move helpers.Move, c *Client) {
	index := len(r.Game.Moves)
	isEven := index%2 == 0
	// for the white player the current move number must be odd,
	// for the black player - even
	if (!isEven && c.User.Id == r.Game.WhiteId) ||
		(isEven && c.User.Id == r.Game.BlackId) {
		if r.Game.HandleMove(move) {
			r.broadcast(UPDATE_GAME)
		}
	}
}
