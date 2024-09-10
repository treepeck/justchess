package game

import (
	"chess-api/models"
	"chess-api/ws/event"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type gameClient struct {
	user             models.User
	conn             *websocket.Conn
	manager          *GameManager
	writeEventBuffer chan event.GameEvent
}

func newClient(conn *websocket.Conn, m *GameManager, user models.User) *gameClient {
	return &gameClient{
		user:             user,
		conn:             conn,
		manager:          m,
		writeEventBuffer: make(chan event.GameEvent),
	}
}
