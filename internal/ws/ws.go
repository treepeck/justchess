// Package ws implements the WebSocket server.
package ws

import (
	"log"
	"math/rand/v2"
	"net/http"

	"justchess/internal/db"
	"justchess/internal/event"
	"justchess/internal/game"
	"justchess/internal/randgen"

	"github.com/gorilla/websocket"
	"github.com/treepeck/chego"
)

const (
	msgNotFound = "Connection will be closed: provided id is not valid"
	msgTooMany  = "There are too many active players. Please, try again later"
	msgConflict = "Please close any previous tabs and reload the page to reconnect"

	// Max number of clients per room or queue.
	clientsThreshold = 100
)

// upgrader is used to establish a WebSocket connection.
// It is safe for concurrent use.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type findRoom struct {
	id  string
	res chan chan *client
}

type createRoom struct {
	id   string
	game game.Game
	res  chan struct{}
}

// Service manages the [room] lifecycle (creation and deletion) and handles
// incomming handshake requests.
type Service struct {
	gameRepo   db.GameRepo
	playerRepo db.PlayerRepo
	rooms      map[string]room
	queues     map[string]queue
	find       chan findRoom
	create     chan createRoom
	remove     chan string
}

func NewService(gr db.GameRepo, pr db.PlayerRepo) Service {
	s := Service{
		gameRepo:   gr,
		playerRepo: pr,
		rooms:      make(map[string]room),
		queues:     make(map[string]queue),
		find:       make(chan findRoom),
		create:     make(chan createRoom, 10),
		remove:     make(chan string, 10),
	}

	controls := [9]struct{ control, bonus int }{{60, 0}, {120, 1}, {180, 0}, {180, 2}, {300, 0}, {300, 2}, {600, 0}, {600, 10}, {900, 10}}
	var i byte
	for i = range 9 {
		q := newQueue(s.create, controls[i].control, controls[i].bonus, gr, pr)
		go q.listenEvents()
		s.queues[string(i+'0')] = q
	}
	return s
}

func (s Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /ws", s.handshake)
	mux.HandleFunc("POST /play-vs-engine", s.createEngineRoom)
}

func (s Service) ListenEvents() {
	for {
		select {
		case e := <-s.create:
			s.createRoom(e)
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
	e := findRoom{
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
	c.send <- event.JSON(event.Error, msgNotFound)
	c.conn.Close()
}

func (s Service) createEngineRoom(rw http.ResponseWriter, r *http.Request) {
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

	id := randgen.GenId(randgen.IdLen)
	var c chego.Color
	if rand.IntN(2) == 1 {
		c = chego.ColorBlack
	}

	g, err := game.SpawnEngineGame(id, p.Id, c, s.gameRepo)
	if err != nil {
		http.Error(rw, msgRoomCreationFailed, http.StatusInternalServerError)
		return
	}
	e := createRoom{
		id:   id,
		game: g,
		res:  make(chan struct{}, 1),
	}
	s.create <- e
	// Wait for response to redirect clients only after room is ready.
	<-e.res

	http.Redirect(rw, r, "/game/engine/"+id, http.StatusFound)
}

func (s Service) createRoom(e createRoom) {
	log.Printf("room %s created", e.id)
	r := newRoom(e.game)
	go r.listenEvents(e.id, s.remove)
	s.rooms[e.id] = r
	e.res <- struct{}{}
}

func (s Service) handleRemoveRoom(id string) {
	if _, exists := s.rooms[id]; exists {
		log.Printf("room %s doesn't exist", id)
		return
	}

	log.Printf("room %s removed", id)
	delete(s.rooms, id)
}
