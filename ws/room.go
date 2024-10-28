package ws

import (
	"chess-api/models/game"
	"chess-api/models/game/enums"
	"chess-api/models/game/helpers"
	"chess-api/models/game/pieces"
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
	Id         uuid.UUID
	game       *game.G
	owner      user.U
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

		case <-r.close:
			return // exit loop
		}
	}
}

// addClient adds a client to the room.
func (r *Room) addClient(c *Client) {
	// if the game hasn`t started yet.
	if r.game.Status == enums.Waiting {
		r.clients[c] = true
	} else {
		// handle reconnections ONLY and deny any other clients to connect.
		if r.game.WhiteId == c.User.Id ||
			r.game.BlackId == c.User.Id {
			r.clients[c] = true
		} else {
			return
		}
	}
	c.currentRoom = r
	r.startGame()
	r.broadcastGameInfo()
}

// removeClient deletes the client from the room and deletes the room itself if the
// room owner disconnects and the game has not been started yet (game aborted).
func (r *Room) removeClient(c *Client) {
	delete(r.clients, c)
	c.currentRoom = nil
	if (r.game.Status == enums.Waiting && r.owner.Id == c.User.Id) ||
		len(r.clients) == 0 {
		c.manager.remove <- r
	}
}

// startGame creates a new game if all clients are connected.
func (r *Room) startGame() {
	if r.game.Status != enums.Waiting || len(r.clients) != 2 {
		return
	}

	r.game.Status = enums.Continues
	// randomize side selection
	players := make([]*Client, 0)
	for c := range r.clients {
		players = append(players, c)
	}

	if rand.Intn(100) < 50 {
		r.game.WhiteId = players[0].User.Id
		r.game.BlackId = players[1].User.Id
	} else {
		r.game.WhiteId = players[1].User.Id
		r.game.BlackId = players[0].User.Id
	}
	r.game.PlayerTurn = r.game.WhiteId // white moves first
}

// broadcastGameInfo broadcasts 4 messages that contains:
//  1. Game status;
//  2. Board (is the game has been started);
//  3. Player`s valid moves (if the game is not over);
//  4. Moves history (is the game has been started).
func (r *Room) broadcastGameInfo() {
	r.broadcastStatus()
	if r.game.Status != enums.Waiting {
		r.broadcastUpdateBoard()
		if r.game.Status == enums.Continues {
			r.broadcastValidMoves()
		}
		r.broadcastMoveHistory()
	}
}

// broadcastUpdateBoard broadcasts the updated board state to players.
func (r *Room) broadcastUpdateBoard() {
	fn := slog.String("func", "broadcaseUpdateBoard")

	// since the map with the struct key cannot be serialized,
	// convert it to map[string]Piece.
	pieces := make(map[string]pieces.Piece)
	for pos, piece := range r.game.Pieces {
		pieces[pos.String()] = piece
	}
	p, err := json.Marshal(pieces)
	if err != nil {
		slog.Warn("cannot Marshal board state", fn, "err", err)
		return
	}
	e := Event{
		Action:  UPDATE_BOARD,
		Payload: p,
	}
	for c := range r.clients {
		c.writeEventBuffer <- e
	}
}

// broadcatValidMoves broadcasts the player`s valid moves for the current turn.
func (r *Room) broadcastValidMoves() {
	fn := slog.String("func", "broadcastValidMoves")

	var p []byte
	var err error
	for c := range r.clients {
		ppm := make([]helpers.PossibleMove, 0)
		if c.User.Id == r.game.PlayerTurn {
			// convert map to slice
			for pm := range r.game.Cvm {
				ppm = append(ppm, pm)
			}
		} else {
			for pm := range r.game.Epm {
				ppm = append(ppm, pm)
			}
		}
		p, err = json.Marshal(ppm)
		if err != nil {
			slog.Warn("cannot Marshal possible moves", fn, "err", err)
			return
		}
		e := Event{
			Action:  VALID_MOVES,
			Payload: p,
		}
		c.writeEventBuffer <- e
	}
}

// broadcastMoveHistory broadcasts the list of the completed moves.
func (r *Room) broadcastMoveHistory() {
	fn := slog.String("func", "broadcastMoveHistory")

	for c := range r.clients {
		p, err := json.Marshal(r.game.Moves)
		if err != nil {
			slog.Warn("cannot marshal move history", fn, "err", err)
			return
		}

		e := Event{
			Action:  MOVES,
			Payload: p,
		}
		c.writeEventBuffer <- e
	}
}

func (r *Room) broadcastStatus() {
	gameDTO := struct {
		White  uuid.UUID    `json:"white"`
		Black  uuid.UUID    `json:"black"`
		Status enums.Status `json:"status"`
	}{
		White:  r.game.WhiteId,
		Black:  r.game.BlackId,
		Status: r.game.Status,
	}

	for c := range r.clients {
		p, _ := json.Marshal(gameDTO)
		e := Event{
			Action:  STATUS,
			Payload: p,
		}
		c.writeEventBuffer <- e
	}
}

func (r *Room) handleTakeMove(move helpers.Move, c *Client) {
	if r.game.PlayerTurn != c.User.Id {
		return
	}

	if r.game.HandleMove(move) {
		r.broadcastGameInfo()
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
