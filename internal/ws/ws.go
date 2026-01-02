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
	CheckOrigin:     func(r *http.Request) bool { return true },
}

const (
	// Declaration of error messages.
	msgInternalError string = "The connection coudn't be established. Please reload the page"
	msgUnauthorized  string = "Sign in to start playing"
	msgNotFound      string = "Requested game does not exist"

	hubId string = "hub"
)

type metadata struct {
	clientId string
	roomId   string
}

type handshake struct {
	r        *http.Request
	rw       http.ResponseWriter
	clientId string
	ch       chan struct{}
}

type Service struct {
	repo       db.Repo
	register   chan handshake
	unregister chan *client
	forward    chan clientEvent
	clients    map[*client]metadata
	rooms      map[string]chan clientEvent
}

func NewService(r db.Repo) Service {
	rooms := make(map[string]chan clientEvent)
	hubCh := make(chan clientEvent)
	rooms[hubId] = hubCh

	hub := newRoom()
	go hub.handle(hubCh)

	return Service{
		repo:       r,
		register:   make(chan handshake),
		unregister: make(chan *client),
		forward:    make(chan clientEvent),
		clients:    make(map[*client]metadata),
		rooms:      rooms,
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

	h := handshake{rw: rw, r: r, ch: make(chan struct{}), clientId: p.Id}

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
	if !exists {
		return
	}

	conn, err := upgrader.Upgrade(h.rw, h.r, nil)
	if err != nil {
		return
	}

	c := newClient(s.unregister, s.forward, conn)
	s.clients[c] = metadata{clientId: h.clientId, roomId: roomId}

	go c.read()
	go c.write()

	// Notify the room that client has joined.
	roomCh <- clientEvent{
		Action:  actionJoin,
		Payload: []byte(h.clientId),
		sender:  c,
	}
}

func (s Service) handleUnregister(c *client) {
	m, exists := s.clients[c]
	if !exists {
		log.Print("client does not exist")
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
		Payload: []byte(m.clientId),
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
		if m.roomId != hubId {
			return
		}

	case actionLeaveMatchmaking:
		if m.roomId != hubId {
			return
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
