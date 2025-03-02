package ws

import (
	"justchess/pkg/game/bitboard"
	"justchess/pkg/game/enums"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type client struct {
	id         uuid.UUID
	room       *Room
	hub        *Hub
	send       chan []byte
	connection *websocket.Conn
}

func newClient(id uuid.UUID, conn *websocket.Conn) *client {
	return &client{
		id:         id,
		connection: conn,
		send:       make(chan []byte),
		hub:        nil,
		room:       nil,
	}
}

func (c *client) readPump() {
	defer func() {
		c.cleanup()
	}()

	for {
		msgType, msg, err := c.connection.ReadMessage()
		if err != nil {
			return
		}

		if msgType == websocket.BinaryMessage {
			c.handleMessage(msg)
		}
	}
}

func (c *client) writePump() {
	defer func() {
		c.cleanup()
	}()

	for {
		msg, ok := <-c.send
		if !ok {
			return
		}

		if err := c.connection.WriteMessage(websocket.BinaryMessage, msg); err != nil {
			return
		}
	}
}

func (c *client) handleMessage(msg []byte) {
	switch msg[len(msg)-1] {
	case CREATE_ROOM:
		c.handleCreateRoom(msg)

	case MAKE_MOVE:
		c.handleMakeMove(msg)
	}
}

func (c *client) handleCreateRoom(msg []byte) {
	if len(msg) != 3 {
		return
	}

	r := NewRoom(c.id, msg[0], msg[1])
	c.hub.add <- r
}

func (c *client) handleMakeMove(msg []byte) {
	if len(msg) != 4 {
		return
	}
	c.room.move <- bitboard.NewMove(int(msg[0]), int(msg[1]), enums.MoveType(msg[2]))
}

func (c *client) cleanup() {
	c.connection.Close()
	if c.room != nil {
		c.room.unregister <- c
	}
	if c.hub != nil {
		c.hub.unregister <- c
	}
}
