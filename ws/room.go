package ws

import (
	"chess-api/models"
	"chess-api/models/enums"
	"chess-api/models/helpers"
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
	Game       *models.Game  `json:"game"`
	Owner      models.User   `json:"owner"`
	Bonus      uint          `json:"bonus"`
	Control    enums.Control `json:"control"`
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
}

// CreateRoomDTO provides necessary data to register a new Room.
type CreateRoomDTO struct {
	Control enums.Control `json:"control"`
	Bonus   uint          `json:"bonus"`
	Owner   models.User   `json:"owner"`
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
	whiteId := uuid.Nil
	blackId := uuid.Nil

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

	r.Game = models.NewGame(r.Id, r.Control, r.Bonus, whiteId, blackId)
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

func (r *Room) handleTakeMove(move helpers.MoveDTO, c *Client) {
	index := r.Game.Moves.Depth() + 1
	isEven := index%2 == 0
	// for the white player the current move number must be odd
	if !isEven && c.User.Id == r.Game.WhiteId {
		if r.Game.TakeMove(move) {
			r.broadcast(UPDATE_GAME)
		}
	} else if isEven && c.User.Id == r.Game.BlackId {
		if r.Game.TakeMove(move) {
			r.broadcast(UPDATE_GAME)
		}
	}
}

func (r *Room) handleGetGame(c *Client) {
	fn := slog.String("func", "handleGetGame")
	// send updated game info back to the client
	p, err := json.Marshal(r.Game)
	if err != nil {
		slog.Warn("cannot Marshal game", fn, "err", err)
		return
	}
	e := Event{
		Action:  UPDATE_GAME,
		Payload: p,
	}
	c.writeEventBuffer <- e
}
