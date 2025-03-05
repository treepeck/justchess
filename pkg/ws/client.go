package ws

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// The client is a middleman between the frontend and the hub.
// Reding and writing messages occurs through the client`s concurrent routines.
type client struct {
	id         uuid.UUID
	hub        *Hub
	room       *Room
	send       chan []byte
	connection *websocket.Conn
}

func newClient(id uuid.UUID, conn *websocket.Conn) *client {
	return &client{
		id:         id,
		send:       make(chan []byte),
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
		if err != nil || data.TimeControl < 1 || data.TimeBonus < 1 {
			return
		}

		r := newRoom(c.hub, c.id, data.TimeControl, data.TimeBonus)
		c.hub.add(r)

	case MAKE_MOVE:

	}
}

func (c *client) cleanup() {
	c.connection.Close()
	if c.hub != nil {
		c.hub.unregister(c)
	}
	if c.room != nil {
		c.room.unregister <- c
	}
}
