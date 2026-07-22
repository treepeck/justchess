package ws

import (
	"errors"
	"github.com/gorilla/websocket"
	"justchess/internal/auth"
	"justchess/internal/response"
	"justchess/internal/talk"
	"log"
	"net/http"
)

// Max number of clients per [room] or [queue].
const clientsThreshold = 100

var (
	errAlreadyRegistered = errors.New("Already connected to a room or queue")
)

// upgrader is used to establish a WebSocket connection.
// It is safe for concurrent use.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type registrant struct {
	id  string
	req *http.Request
	res http.ResponseWriter
	err chan error
}

type findRegistrant struct {
	id  string
	res chan chan registrant
}

type createRoom struct {
	id       string
	channels talk.GameChannels
	res      chan chan registrant
}

// Service manages the [room] lifecycle (creation and deletion) and handles
// incomming handshake requests.
type Service struct {
	find    chan talk.GameFinder
	findReg chan findRegistrant
	create  chan createRoom
	destroy chan string
	rooms   map[string]room
	queues  map[string]queue
}

var timeControls = []struct {
	c int // Control.
	b int // Bonus.
}{{1, 0}, {2, 1}, {3, 0}, {3, 2}, {5, 0}, {5, 2}, {10, 0}, {10, 10}, {15, 10}}

func NewService(find chan talk.GameFinder, create chan talk.GameCreator) Service {
	// Declare a distinct queue for each available time control.
	queues := make(map[string]queue, 9)
	for _, c := range timeControls {
		queues[string(byte(c.c+'0'))] = initQueue(c.c, c.b, create)
	}

	return Service{
		find:    find,
		findReg: make(chan findRegistrant),
		create:  make(chan createRoom),
		destroy: make(chan string),
		rooms:   make(map[string]room),
		queues:  queues,
	}
}

func (s Service) RegisterRoutes(authService auth.Service, mux *http.ServeMux) {
	mux.HandleFunc("GET /ws/{id}", authService.MustAuthorize(s.handshake))
}

func (s Service) Listen() {
	for {
		select {
		case f := <-s.findReg:
			s.onFind(f)
		case c := <-s.create:
			s.onCreate(c)
		case id := <-s.destroy:
			s.onDestroy(id)
		}
	}
}

func (s Service) onFind(f findRegistrant) {
	if r, exists := s.rooms[f.id]; exists {
		f.res <- r.register
		return
	}
	if q, exists := s.queues[f.id]; exists {
		f.res <- q.register
		return
	}
	f.res <- nil
}

func (s Service) onCreate(c createRoom) {
	log.Printf("room %s created", c.id)
	r := initRoom(c.channels)
	s.rooms[c.id] = r
	c.res <- r.register
}

func (s Service) onDestroy(id string) {
	log.Printf("room %s destroyed", id)
	delete(s.rooms, id)
}

func (s Service) handshake(rw http.ResponseWriter, r *http.Request) {
	session, ok := r.Context().Value(auth.SessionKey).(auth.Session)
	if !ok {
		log.Print("request context is broken")
		http.Error(rw, response.Unauthorized, http.StatusUnauthorized)
		return
	}

	f := findRegistrant{
		id:  r.PathValue("id"),
		res: make(chan chan registrant),
	}
	s.findReg <- f

	register := <-f.res
	if register == nil {
		// If the game room wasn't found, it might need to be created.
		gf := talk.GameFinder{
			Id:  f.id,
			Res: make(chan talk.GameChannels),
		}
		s.find <- gf
		gc := <-gf.Res
		if gc.In == nil || gc.Out == nil || gc.Ban == nil {
			http.Error(rw, response.NotFound, http.StatusNotFound)
			return
		}
		c := createRoom{
			id:       f.id,
			channels: gc,
			res:      make(chan chan registrant),
		}
		s.create <- c
		register = <-c.res
	}
	// If the endpoint was found, wait for response.
	dto := registrant{
		id:  session.Id,
		req: r,
		res: rw,
		err: make(chan error),
	}
	register <- dto
	if err := <-dto.err; err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
	}
}
