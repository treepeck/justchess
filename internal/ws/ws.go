// Package ws implements the WebSocket server. It is a transport-layer package
// that knows about game logic as little as possible. All it does is instantiates
// and maintains incomming connections and broadcasts messages.
//
// This package follows the Pub/Sub-like architecture, where [room] is equivalent
// to the Topic, and by regisreting to the room, [client]s will recieve all [message]s
// that are broadcasted in this [room].
//
// [Service] is a central instance of the package. It manages the [room] lifecycle and
// handles incomming requests.
package ws

import (
	"github.com/gorilla/websocket"
	"justchess/internal/auth"
	"justchess/internal/db"
	"justchess/internal/response"
	"net/http"
	"sync"
)

// upgrader is used to establish a WebSocket connection.
// It is safe for concurrent use.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Service struct {
	sync.Mutex
	gameRepo db.GameRepo
	rooms    map[string]room
}

func InitService(gr db.GameRepo) Service {
	return Service{
		gameRepo: gr,
	}
}

func (s Service) RegisterRoutes(as auth.Service, mux *http.ServeMux) {
	mux.HandleFunc("GET /ws/{id}", as.MustAuthorize(s.handshake))
}

// handshake handles a WebSocket handshake request by instantiating
// a [client] and registering it in the named [room].
//
// First of all, the [room] is looked up by the named id in the internal
// [Service] memory. If it's not found, then the database is probed.
// The handshake is declined if there is no [room] or databse record
// with the named id.
func (s Service) handshake(rw http.ResponseWriter, r *http.Request) {
	s.Lock()
	// Check does room with the named id exist.
	r, ok := s.rooms[r.PathValue("id")]
	if !ok {
		http.Error(rw, r, response.NotFound, http.StatusNotFound)
		return
	}
	s.gameRepo.
		s.Unlock()
	// Identify the player.
	p, ok := r.Context().Value(auth.PlayerKey).(db.Player)
	if !ok {
		log.Print("ws: request context is broken")
		http.Error(rw, r, response.InternalError, http.StatusInternalServerError)
		return
	}
	// Instantiate the client.
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		// Upgrader writes the response, so simply return here.
		return
	}
	c := initClient(conn)
	// Register the client in a room.
	r.register(p, c)
}
