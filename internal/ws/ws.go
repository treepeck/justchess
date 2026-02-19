// Package ws implements the WebSocket server.
package ws

import (
	"log"
	"net/http"

	"justchess/internal/db"

	"github.com/gorilla/websocket"
	"github.com/treepeck/chego"
)

const (
	msgNotFound string = "Connection will be closed: provided id is not valid."
	msgConflict string = "Please close any previous tabs and reload the page to reconnect"
)

// upgrader is used to establish a WebSocket connection.
// It is safe for concurrent use.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type createRoom struct {
	id, whiteId, blackId string
	control              control
	res                  chan error
}

type findRegister struct {
	id  string
	res chan chan *client
}

// Service manages the [room] lifecycle (creation and deletion), handles
// incomming handshake requests, and modifies the database (stores completed
// games and updated player's ratings).
type Service struct {
	playerRepo db.PlayerRepo
	gameRepo   db.GameRepo
	create     chan createRoom
	remove     chan string
	find       chan findRegister
	rooms      map[string]*room
	queues     map[string]queue
}

func NewService(pr db.PlayerRepo, gr db.GameRepo) Service {
	s := Service{
		playerRepo: pr,
		gameRepo:   gr,
		create:     make(chan createRoom),
		remove:     make(chan string),
		find:       make(chan findRegister),
		rooms:      make(map[string]*room),
		queues:     make(map[string]queue, 9),
	}

	for i := byte(1); i < 10; i++ {
		q := newQueue(i)
		go q.listenEvents(s.create)
		s.queues[string(i+'0')] = q
	}

	return s
}

// RegisterRoute registers the handshake enpoint to the specified ServeMux.
func (s Service) RegisterRoute(mux *http.ServeMux) {
	mux.HandleFunc("/ws", s.handshake)
}

func (s Service) ListenEvents() {
	for {
		select {
		case e := <-s.create:
			s.handleCreateRoom(e)
		case id := <-s.remove:
			s.handleRemoveRoom(id)
		case e := <-s.find:
			if q, exist := s.queues[e.id]; exist {
				e.res <- q.register
			} else if r, exist := s.rooms[e.id]; exist {
				e.res <- r.register
			} else {
				e.res <- nil
			}
		}
	}
}

// handshake handles WebSocket handshake requests.  Each incoming request must
// include an 'id' parameter that identifies the room or queue the client is
// attempting to join.  The request will be denied if the session cookie is
// missing or expired.
//
// An error event will be sent to the client immediately after the connection
// is opened in the following cases:
//   - No room or queue exists with the provided id;
//   - The client is already registered in the room or queue.
//
// The connection will be closed after the error event is sent.
func (s Service) handshake(rw http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("Auth")
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	p, err := s.playerRepo.SelectBySessionId(session.Value)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")

	// Create WebSocket connection.
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		// Simply return here since the upgrader writes the response.
		return
	}
	c := newClient(conn, p)
	go c.read()
	go c.write()

	// Search for a room or queue with the given id.
	e := findRegister{
		id:  id,
		res: make(chan chan *client, 1),
	}
	s.find <- e

	// Handle response.
	if register := <-e.res; register != nil {
		register <- c
		return
	}

	// Send error event to the client.
	if raw, err := newEncodedEvent(actionError, msgNotFound); err == nil {
		c.send <- raw
	}
	c.conn.Close()
}

func (s Service) handleCreateRoom(e createRoom) {
	err := s.gameRepo.Insert(e.id, e.whiteId, e.blackId, e.control.minutes, e.control.bonus)
	defer func() { e.res <- err }()
	if err != nil {
		return
	}

	log.Printf("room %s created", e.id)

	r := newRoom(e.id, e.whiteId, e.blackId, e.control.minutes, e.control.bonus)
	go r.listenEvents(s.remove)

	s.rooms[e.id] = r
}

func (s Service) handleRemoveRoom(id string) {
	r, exist := s.rooms[id]
	if !exist {
		return
	}

	// Encode moves.
	indices := make([]byte, len(r.moves))
	for i, m := range r.moves {
		indices[i] = m.index
	}
	encoded := chego.HuffmanEncoding(indices)

	err := s.gameRepo.Update(r.game.Result, r.game.Termination, len(r.moves), encoded, id)
	if err != nil {
		log.Print(err)
	}

	log.Printf("room %s removed", id)
	delete(s.rooms, id)
}
