package game

import (
	"chess-api/repository"
	"log/slog"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	WriteBufferSize: 1024,
	ReadBufferSize:  1024,
	CheckOrigin: func(req *http.Request) bool {
		// TODO: change to the client (front-end) domain later
		return true
	},
}

type GameManager struct {
	sync.Mutex
	blackPlayer *gameClient
	whitePlayer *gameClient
}

func NewManager() *GameManager {
	return &GameManager{
		blackPlayer: nil,
		whitePlayer: nil,
	}
}

func (m *GameManager) HandleConnection(rw http.ResponseWriter, r *http.Request) {
	fn := slog.String("func", "HandleConnection")

	idStr := r.URL.Query().Get("id")
	userId, err := uuid.Parse(idStr)
	if err != nil {
		slog.Warn("cannot parse uuid", fn, "err", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	u := repository.FindUserById(userId)
	if u == nil {
		slog.Warn("user not found", fn, "err", err)
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err = upgrader.Upgrade(rw, r, nil)
	if err != nil {
		slog.Warn("error while upgrading the connection", fn, "err", err)
		return
	}

	// c := newClient(conn, m, *u)
	// m.addClient(c)
}
