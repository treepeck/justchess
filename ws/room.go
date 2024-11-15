package ws

import (
	"chess-api/models/game"
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
	"chess-api/models/user"
	"encoding/json"
	"log/slog"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// Room stores players and a game.
// There is always one single Room for the every game.
type Room struct {
	Id         uuid.UUID
	game       *game.G
	owner      user.U
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	close      chan bool // channel to break the room run loop.
}

// CreateRoomDTO provides necessary data to register a new Room.
type CreateRoomDTO struct {
	Control enums.Control `json:"control"`
	Bonus   uint          `json:"bonus"`
}

// newRoom creates and runs a new room.
func newRoom(cr CreateRoomDTO, owner *Client) *Room {
	r := &Room{
		Id:         uuid.New(),
		owner:      owner.User,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		close:      make(chan bool),
	}
	// create a new game without players.
	r.game = game.NewG(r.Id, cr.Control, cr.Bonus, uuid.Nil, uuid.Nil)
	go r.run()
	return r
}

func (r *Room) run() {
	for {
		select {
		case c := <-r.register:
			r.addClient(c)

		case c := <-r.unregister:
			r.removeClient(c)

		case <-r.game.White.Ticker.C:
			r.handleWhiteTimeTick()

		case <-r.game.Black.Ticker.C:
			r.handleBlackTimeTick()

		case <-r.close:
			return // exit loop
		}
	}
}

// addClient adds a client to the room.
func (r *Room) addClient(c *Client) {
	fn := slog.String("func", "room.addClient")
	slog.Debug("client "+c.User.Name+" added", fn)

	switch r.game.Status {
	// deny any connections
	case enums.Aborted, enums.Over:
		return

	case enums.Waiting:
		r.clients[c] = true
		c.currentRoom = r
		r.startGame()

	case enums.Leave:
		// if the client was connected before
		if r.game.Black.Id == c.User.Id {
			r.clients[c] = true
			c.currentRoom = r
			r.resumeGame(enums.Black)
		} else if r.game.White.Id == c.User.Id {
			r.clients[c] = true
			c.currentRoom = r
			r.resumeGame(enums.White)
		}
	}
}

// removeClient deletes the client from the room and deletes the room itself
// if the there is no one left in the room.
func (r *Room) removeClient(c *Client) {
	fn := slog.String("func", "room.removeClient")
	slog.Debug("client "+c.User.Name+" removed", fn)

	delete(r.clients, c)
	c.currentRoom = nil

	if len(r.clients) == 0 {
		c.manager.remove <- r
		return
	}

	if r.game.Status == enums.Continues {
		r.game.Status = enums.Leave
		if c.User.Id == r.game.White.Id {
			r.game.White.ExtraTime = 20 * time.Second
			r.game.White.Ticker.Reset(time.Second)
		} else if c.User.Id == r.game.Black.Id {
			r.game.Black.ExtraTime = 20 * time.Second
			r.game.Black.Ticker.Reset(time.Second)
		}
	}
}

// startGame creates a new game if all clients are connected.
func (r *Room) startGame() {
	if len(r.clients) != 2 {
		return
	}
	// randomly generate players sides
	players := make([]*Client, 0)
	for c := range r.clients {
		players = append(players, c)
	}
	var whiteId, blackId uuid.UUID
	if rand.Intn(100) < 50 {
		whiteId = players[0].User.Id
		blackId = players[1].User.Id
	} else {
		whiteId = players[1].User.Id
		blackId = players[0].User.Id
	}
	r.game.StartGame(whiteId, blackId)
	// broadcast game info
	for c := range r.clients {
		c.sendEvent(GAME_INFO, r.game)
		c.sendValidMoves(r.game.CurrentValidMoves)
	}
}

func (r *Room) handleWhiteTimeTick() {
	if r.game.CurrentTurn == enums.White {
		r.game.White.DecrementTime()
	}
	if !r.game.White.IsConnected || len(r.game.Moves) == 0 {
		r.game.White.DecrementExtraTime()
	}
	// handle timeouts
	if r.game.White.Time == 0 {
		r.endGame(enums.Timeout, int(enums.Black))
	} else if r.game.White.ExtraTime == 0 {
		switch r.game.Status {
		case enums.Continues:
			r.abortGame()

		case enums.Leave:
			r.endGame(enums.Resignation, int(enums.Black))
		}
	}
}

func (r *Room) handleBlackTimeTick() {
	if r.game.CurrentTurn <= enums.Black {
		r.game.Black.DecrementTime()
	}
	if !r.game.Black.IsConnected || len(r.game.Moves) == 1 {
		r.game.Black.DecrementExtraTime()
	}
	// handle timeouts
	if r.game.Black.Time == 0 {
		r.endGame(enums.Timeout, int(enums.White))
	} else if r.game.Black.ExtraTime == 0 {
		switch r.game.Status {
		case enums.Continues:
			r.abortGame()

		case enums.Leave:
			r.endGame(enums.Resignation, int(enums.White))
		}
	}
}

func (r *Room) abortGame() {
	r.game.Status = enums.Aborted
	r.game.White.Ticker.Stop()
	r.game.Black.Ticker.Stop()
	for c := range r.clients {
		c.writeEventBuffer <- Event{
			Action:  ABORT,
			Payload: nil,
		}
	}
}

func (r *Room) resumeGame(side enums.Color) {
	r.game.Status = enums.Continues

	if side == enums.White {
		r.game.White.IsConnected = true
		if r.game.CurrentTurn == enums.Black {
			r.game.White.Ticker.Stop()
		}
	} else {
		r.game.Black.IsConnected = true
		if r.game.CurrentTurn == enums.White {
			r.game.Black.Ticker.Stop()
		}
	}

	for c := range r.clients {
		c.sendEvent(GAME_INFO, r.game)
	}
}

// endGame writes the game data to the db and
// removes the players from the room.
func (r *Room) endGame(res enums.GameResult, w int) {
	// repository.SaveGame(r.game)
	r.game.EndGame(res, w)
	// broadcast game result
	for c := range r.clients {
		endGameDTO := struct {
			Result enums.GameResult `json:"r"`
			Winner int              `json:"w"`
		}{
			Result: res, Winner: w,
		}

		p, _ := json.Marshal(endGameDTO)
		e := Event{
			Action:  END_RESULT,
			Payload: p,
		}
		c.writeEventBuffer <- e
	}
}

// handleTakeMove handles player`s moves.
func (r *Room) handleTakeMove(move helpers.Move, c *Client) {
	// ignore moves if it is not a player`s turn
	if (c.User.Id == r.game.White.Id && r.game.CurrentTurn != enums.White) ||
		(c.User.Id == r.game.Black.Id && r.game.CurrentTurn != enums.Black) {
		return
	}

	if r.game.HandleMove(&move) {
		for c := range r.clients {
			c.sendEvent(LAST_MOVE, move)
			c.sendValidMoves(r.game.CurrentValidMoves)
		}
		if r.game.Status == enums.Over {
			r.endGame(r.game.Result, r.game.Winner)
		}
	}
}

// MarshalJSON serializes room into json string.
func (r *Room) MarshalJSON() ([]byte, error) {
	roomDTO := struct {
		Id      uuid.UUID     `json:"id"`
		Control enums.Control `json:"control"`
		Bonus   uint          `json:"bonus"`
		Owner   user.U        `json:"owner"`
	}{
		Id:      r.Id,
		Control: r.game.Control,
		Bonus:   r.game.Bonus,
		Owner:   r.owner,
	}
	return json.Marshal(roomDTO)
}
