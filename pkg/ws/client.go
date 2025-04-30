package ws

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// client is a middleman between the frontend and server.
// Reading and writing messages occurs through the client's concurrent routines.
type client struct {
	id      uuid.UUID
	name    string
	isGuest bool
	hub     *Hub
	room    *Room
	// send channel must be a buffered one, otherwise if the routine writes to it but the client
	// drops connection, the routine will wait forever.
	send       chan []byte
	connection *websocket.Conn
}

func newClient(id uuid.UUID, name string, isGuest bool, conn *websocket.Conn) *client {
	return &client{
		id:         id,
		name:       name,
		isGuest:    isGuest,
		send:       make(chan []byte, 256),
		connection: conn,
	}
}

func (c *client) readRoutine() {
	defer func() {
		c.cleanup()
	}()

	c.connection.SetReadLimit(maxMessageSize)
	c.connection.SetReadDeadline(time.Now().Add(pongWait))
	c.connection.SetPongHandler(func(appData string) error { return c.connection.SetReadDeadline(time.Now().Add(pongWait)) })

	for {
		_, msg, err := c.connection.ReadMessage()
		if err != nil {
			return
		}

		c.handleMessage(msg)
	}
}

func (c *client) writeRoutine() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.cleanup()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.connection.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
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
		data := CreateRoomDTO{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil || data.TimeControl < 1 || data.TimeBonus < 0 || c.hub == nil ||
			(!data.IsVSEngine && c.isGuest) {
			return
		}

		r := newRoom(c.hub, c.name, data.IsVSEngine, data.TimeControl, data.TimeBonus)
		c.hub.add(r)

	case MAKE_MOVE:
		var move MoveDTO
		err = json.Unmarshal(msg.Data, &move)
		if c.room == nil || err != nil {
			return
		}
		c.room.move <- moveEvent{client: c, move: move}

	case CHAT:
		data := ChatDTO{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil || c.room != nil {
			c.room.chat <- chatEvent{data: data, client: c}
		}

	case RESIGN:
		if c.room != nil {
			c.room.clientEvents <- clientEvent{client: c, eType: typeResign}
		}

	case DRAW_OFFER:
		if c.room != nil {
			c.room.clientEvents <- clientEvent{client: c, eType: typeOfferDraw}
		}

	case DECLINE_DRAW:
		if c.room != nil {
			c.room.clientEvents <- clientEvent{client: c, eType: typeDeclineDraw}
		}
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
