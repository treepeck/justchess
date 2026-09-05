package ws

import (
	"github.com/gorilla/websocket"
	"justchess/internal/randgen"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Service is a HTTP server that serves a single specific endpoint called "/handshake".
// Handshake is a special HTTP request that the client platform sends to establish
// the WebSocket connection. Such request must have a specific format, defined
// in RFC 6455. Request validation along with protocol switching are handled
// entirely by the "gorilla/websocket" package. After it does it's job, the [Service]
// stores the connection object. All subsequent interaction between the client and
// [Service] occurs outside the scope of the handshake handler.
type Service struct {
	clients map[string]*client
}

func NewService() Service {
	return Service{
		clients: make(map[string]*client),
	}
}

func (s Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /handshake", s.handshake)
}

func (s Service) handshake(rw http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Printf("error while trying to upgrade the connection: %v\n", err)
		return
	}

	tmpId := randgen.GenId(randgen.IdLen)
	c := initClient(tmpId, conn)
	s.clients[tmpId] = c
}
