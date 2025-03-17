package ws

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// client is a middleman between the frontend and server.
// Reding and writing messages occurs through the client's concurrent routines.
type client struct {
	id   uuid.UUID
	name string
	hub  *Hub
	room *Room
	// send channel must be a buffered one, otherwise if the routine writes to it but the client
	// drops connection, the routine will wait forever.
	send       chan []byte
	connection *websocket.Conn
}

func newClient(id uuid.UUID, name string, conn *websocket.Conn) *client {
	return &client{
		id:         id,
		name:       name,
		send:       make(chan []byte, 256),
		connection: conn,
	}
}

func (c *client) readRoutine() {
	defer func() {
		c.cleanup()
	}()

	for {
		_, msg, err := c.connection.ReadMessage()
		if err != nil {
			return
		}

		c.handleMessage(msg)
	}
}

func (c *client) writeRoutine() {
	defer func() {
		c.cleanup()
	}()

	for {
		msg, ok := <-c.send
		if !ok {
			return
		}

		if err := c.connection.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *client) handleMessage(raw []byte) {
	msg := Message{}
	err := json.Unmarshal(raw, &msg)
	if err != nil {
		return
	}

	switch msg.Type {
	case CREATE_ROOM:
		data := CreateRoomData{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil || data.TimeControl < 1 || data.TimeBonus < 0 || c.hub == nil {
			return
		}

		r := newRoom(c.hub, c.name, data.IsVSEngine, data.TimeControl, data.TimeBonus)

		c.hub.add(r)

	case MAKE_MOVE:
		data := MoveData{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil || c.room == nil {
			return
		}

		c.room.handle(data, c)

	case CHAT:
		data := ChatData{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil || c.room == nil {
			return
		}

		c.room.broadcastChat(data, c)

	case RESIGN:
		if c.room == nil {
			return
		}
		c.room.handleResign(c.name)
	}
}

func (c *client) cleanup() {
	c.connection.Close()
	if c.hub != nil {
		c.hub.unregister(c)
	}
	if c.room != nil {
		c.room.unregister(c)
	}
}
