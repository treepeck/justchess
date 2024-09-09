package ws

import (
	"chess-api/models"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type client struct {
	user             models.User
	conn             *websocket.Conn
	manager          *Manager
	writeEventBuffer chan event
	currentRoom      *room
}

func newClient(conn *websocket.Conn, m *Manager, user models.User) *client {
	return &client{
		user:             user,
		conn:             conn,
		manager:          m,
		writeEventBuffer: make(chan event),
		currentRoom:      nil,
	}
}

// Reads all incoming messages from the connection.
func (c *client) readEvents() {
	fn := slog.String("func", "readEvents")

	defer c.manager.removeClient(c)

	// set the read deadline to limit inactive connections
	c.conn.SetReadLimit(10000)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		slog.Warn("error while setting the read deadline", fn, "err", err)
		return
	}
	c.conn.SetPongHandler(c.pongHandler)

	for {
		_, data, err := c.conn.ReadMessage()
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway,
			websocket.CloseAbnormalClosure) {
			slog.Warn("error while reading "+
				"messages from a connection", fn, "err", err,
			)
			return
		}

		var e event
		err = json.Unmarshal(data, &e)
		if err != nil {
			slog.Warn("cannot Unmarshal event", fn, "err", err)
			return
		}
		c.handleEvent(e)
	}

}

func (c *client) writeEvents() {
	fn := slog.String("func", "writeEvents")

	defer c.manager.removeClient(c)

	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case e, ok := <-c.writeEventBuffer:
			c.conn.SetWriteDeadline(time.Now().Add(pingInterval))
			if !ok {
				slog.Info("connection closed", fn)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, e.marshal()); err != nil {
				slog.Warn("failed to write event", fn, "err", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(pingInterval))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte("Ping")); err != nil {
				return
			}
		}
	}
}

func (c *client) pongHandler(_ string) error {
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}

func (c *client) handleEvent(e event) {
	fn := slog.String("func", "handleEvent")

	switch e.Action {
	case GET_ROOMS:
		rooms, err := json.Marshal(c.manager.getAllRooms())
		if err != nil {
			slog.Warn("cannot Marshal rooms", fn, "err", err)
			return
		}
		e.Action = UPDATE_ROOMS
		e.Payload = rooms
		c.writeEventBuffer <- e

	case CREATE_ROOM:
		if c.currentRoom != nil {
			slog.Warn("cannot create room, already joined", fn)
			return
		}

		var cr CreateRoomDTO
		err := json.Unmarshal(e.Payload, &cr)
		if err != nil {
			slog.Debug("canot unmarshal CreateRoomDTO", fn, "err", err)
			return
		}

		c.manager.createRoom(cr)

	case JOIN_ROOM:
		if c.currentRoom != nil {
			slog.Debug("cannot join room, already joined", fn)
			return
		}

		if r := c.findRoomByPayload(e.Payload); r != nil {
			c.joinRoom(r)
		}

	case LEAVE_ROOM:
		if c.currentRoom == nil {
			slog.Debug("doesnt connected to a room", fn)
			return
		}

		c.leaveRoom()
	}
}

func (c *client) findRoomByPayload(payload json.RawMessage) *room {
	fn := slog.String("func", "findRoomByPayload")

	var idStr string
	json.Unmarshal(payload, &idStr)
	roomId, err := uuid.Parse(idStr)
	if err != nil {
		slog.Warn("cannot parse roomId ", fn, "err", err)
		return nil
	}
	if r := c.manager.findRoomById(roomId); r != nil {
		return r
	} else {
		slog.Debug("room not found", fn)
		return nil
	}
}

func (c *client) leaveRoom() {
	if c.currentRoom != nil {
		c.currentRoom.remove <- c

		if len(c.currentRoom.clients) < 1 {
			c.manager.removeRoom(c.currentRoom)
		}

		c.currentRoom = nil
	}
}

func (c *client) joinRoom(r *room) {
	r.add <- c
	c.currentRoom = r

	p, _ := json.Marshal(r.Id)
	e := event{
		Action:  CHANGE_ROOM,
		Payload: p,
	}
	c.writeEventBuffer <- e
}
