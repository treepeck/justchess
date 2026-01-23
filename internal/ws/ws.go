package ws

import (
	"log"
	"net/http"

	"justchess/internal/db"

	"github.com/gorilla/websocket"
)

// upgrader is used to establish a WebSocket connection.
// It is safe for concurrent use.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Declaration of error messages.
const (
	msgInternalError string = "The connection cannot be established. Please reload the page"
	msgUnauthorized  string = "Sign in to start playing"
	msgNotFound      string = "Room doesn't exist"
)

type handshake struct {
	r      *http.Request
	rw     http.ResponseWriter
	player db.Player
	ch     chan struct{}
}

type Service struct {
	repo   db.Repo
	create chan createRoomEvent
	remove chan string
	find   chan findRoomEvent
	rooms  map[string]room
	queues map[string]queue
}

func NewService(r db.Repo) Service {
	s := Service{
		repo:   r,
		create: make(chan createRoomEvent),
		remove: make(chan string),
		find:   make(chan findRoomEvent),
		rooms:  make(map[string]room),
	}

	s.queues = make(map[string]queue, 9)
	// Add queue for each game mode.
	var params = []struct {
		id      string
		control int
		bonus   int
	}{
		{"1", 1, 0},
		{"2", 2, 1},
		{"3", 3, 0},
		{"4", 3, 2},
		{"5", 5, 0},
		{"6", 5, 2},
		{"7", 10, 0},
		{"8", 10, 10},
		{"9", 15, 10},
	}
	for _, param := range params {
		q := newQueue(param.control, param.bonus)
		s.queues[param.id] = q
		// Will run until the program exists.
		go q.listenEvents(s.create)
	}

	return s
}

func (s Service) ListenEvents() {
	for {
		select {
		case e := <-s.create:
			s.createRoom(e)
		case id := <-s.remove:
			s.removeRoom(id)
		case e := <-s.find:
			e.res <- s.rooms[e.id]
		}
	}
}

// RegisterRoutes registers endpoints to the specified mux.
func (s Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /ws", s.serveWS)
}

// Concurrently accepts the WebSocket handshake requests.
func (s Service) serveWS(rw http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("Auth")
	if err != nil {
		http.Error(rw, msgUnauthorized, http.StatusUnauthorized)
		return
	}

	p, err := s.repo.SelectPlayerBySessionId(c.Value)
	if err != nil {
		http.Error(rw, msgUnauthorized, http.StatusUnauthorized)
		return
	}

	h := handshake{rw: rw, r: r, ch: make(chan struct{}), player: p}

	id := h.r.URL.Query().Get("id")
	queue, exists := s.queues[id]
	if exists {
		queue.register <- h
		// Wait for the response.
		<-h.ch
		return
	}

	e := findRoomEvent{
		id:  id,
		res: make(chan room, 1),
	}
	s.find <- e
	room := <-e.res
	if room.register != nil {
		room.register <- h
		// Wait for the response.
		<-h.ch
		return
	}

	http.Error(rw, msgNotFound, http.StatusNotFound)
}

func (s Service) createRoom(e createRoomEvent) {
	err := s.repo.InsertGame(e.id, e.whiteId, e.blackId, e.control, e.bonus)
	defer func() { e.res <- err }()

	if err != nil {
		log.Print(err)
		return
	}

	r := newRoom(e.id, e.whiteId, e.blackId)
	go r.listenEvents(s.remove)
	s.rooms[e.id] = r

	log.Printf("room %s created", e.id)
}

func (s Service) removeRoom(id string) {
	delete(s.rooms, id)
	log.Printf("room %s removed", id)
}
