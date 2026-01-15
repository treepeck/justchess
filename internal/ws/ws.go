package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"justchess/internal/db"
	"justchess/internal/matchmaking"

	"github.com/gorilla/websocket"
)

// upgrader is used to establish a WebSocket connection.
// It is safe for concurrent use.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

const (
	// Declaration of error messages.
	msgInternalError string = "The connection coudn't be established. Please reload the page"
	msgUnauthorized  string = "Sign in to start playing"
	msgNotFound      string = "Requested game does not exist"

	homeId string = "home"
)

type metadata struct {
	player         db.Player
	roomId         string
	isFindingMatch bool
}

type handshake struct {
	r      *http.Request
	rw     http.ResponseWriter
	player db.Player
	ch     chan struct{}
}

type Service struct {
	pool       matchmaking.Pool
	repo       db.Repo
	register   chan handshake
	unregister chan *client
	forward    chan clientEvent
	clients    map[*client]*metadata
	rooms      map[string]chan clientEvent
}

func NewService(r db.Repo) Service {
	return Service{
		repo:       r,
		pool:       matchmaking.NewPool(),
		register:   make(chan handshake),
		unregister: make(chan *client),
		forward:    make(chan clientEvent),
		clients:    make(map[*client]*metadata),
		rooms:      make(map[string]chan clientEvent),
	}
}

// RegisterRoutes registers endpoints to the specified mux.
func (s Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /ws", s.serveWS)
}

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

	s.register <- h
	<-h.ch
}

func (s Service) EventBus() {
	for {
		select {
		case h := <-s.register:
			s.handleRegister(h)

		case c := <-s.unregister:
			s.handleUnregister(c)

		case e := <-s.forward:
			s.forwardEvent(e)
		}
	}
}

func (s Service) handleRegister(h handshake) {
	defer func() { h.ch <- struct{}{} }()

	roomId := h.r.URL.Query().Get("rid")
	roomCh, exists := s.rooms[roomId]
	if !exists && roomId != homeId {
		http.Error(h.rw, msgNotFound, http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(h.rw, h.r, nil)
	if err != nil {
		// upgrader writes the response, so simply return here.
		return
	}

	c := newClient(conn)
	s.clients[c] = &metadata{player: h.player, roomId: roomId}

	go c.read(s.unregister, s.forward)
	go c.write()

	// Notify the room that client has joined.
	if exists {
		roomCh <- clientEvent{
			Action:  actionJoin,
			Payload: []byte(h.player.Id),
			sender:  c,
		}
	}
}

func (s Service) handleUnregister(c *client) {
	m, exists := s.clients[c]
	if !exists {
		log.Print("client does not exist")
		return
	}

	if m.roomId == homeId {
		if m.isFindingMatch {
			s.pool.Leave(m.player.Id, m.player.Rating)
		}
		return
	}

	roomCh, exists := s.rooms[m.roomId]
	if !exists {
		log.Print("room does not exist")
		return
	}

	// Notify the room that client has leaved.
	roomCh <- clientEvent{
		Action:  actionLeave,
		Payload: []byte(m.player.Id),
		sender:  c,
	}
}

func (s Service) forwardEvent(e clientEvent) {
	m, exists := s.clients[e.sender]
	if !exists {
		log.Print("client does not exist")
		return
	}

	switch e.Action {
	case actionJoinMatchmaking:
		if m.roomId == homeId && !m.isFindingMatch {
			var t matchmaking.Ticket
			if err := json.Unmarshal(e.Payload, &t); err != nil {
				log.Print(err)
				return
			}

			m.isFindingMatch = true

			s.pool.Join(m.player.Id, m.player.Rating, t)
		}

	case actionLeaveMatchmaking:
		if m.roomId == homeId && m.isFindingMatch {
			s.pool.Leave(m.player.Id, m.player.Rating)
			m.isFindingMatch = false
		}

	default:
		roomCh, exists := s.rooms[m.roomId]
		if !exists {
			log.Print("room does not exist")
			return
		}

		// Notify the room that client has leaved.
		roomCh <- e
	}
}
