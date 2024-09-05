package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type client struct {
	conn             *websocket.Conn
	manager          *Manager
	writeEventBuffer chan event
	currentRoom      *room
}

func newClient(conn *websocket.Conn, m *Manager) *client {
	return &client{
		conn:             conn,
		manager:          m,
		writeEventBuffer: make(chan event),
		currentRoom:      nil,
	}
}

// Reads all incoming messages from the connection.
func (c *client) readEvents() {
	defer c.manager.removeClient(c)

	// set the read deadline to limit inactive connections
	c.conn.SetReadLimit(10000)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println("readEvents: error while setting the read deadline: ", err)
		return
	}
	c.conn.SetPongHandler(c.pongHandler)

	for {
		_, data, err := c.conn.ReadMessage()
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway,
			websocket.CloseAbnormalClosure) {
			log.Println("readEvents: error while reading "+
				"messages from a connection: ", err,
			)
			return
		}

		var e event
		err = json.Unmarshal(data, &e)
		if err != nil {
			log.Println("readEvents: cannot Unmarshal event", err)
			return
		}
		c.handleEvent(e)
	}

}

func (c *client) writeEvents() {
	defer c.manager.removeClient(c)

	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case e, ok := <-c.writeEventBuffer:
			c.conn.SetWriteDeadline(time.Now().Add(pingInterval))
			if !ok {
				log.Println("writeEvents: connection closed")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, e.marshal()); err != nil {
				log.Println("writeEvents: failed to write event: ", err)
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
	e.Sender = c

	switch e.Action {

	case CREATE_ROOM:
		if c.currentRoom != nil {
			log.Println("handleEvent: cannot create room, already joined")
			return
		}

		r := c.manager.createRoom()
		c.joinRoom(r)

	case JOIN_ROOM:
		if c.currentRoom != nil {
			log.Println("handleEvent: cannot joint room, already joined")
			return
		}

		if r := c.findRoomByPayload(e.Payload); r != nil {
			c.joinRoom(r)
		}

	case LEAVE_ROOM:
		if c.currentRoom == nil {
			log.Println("handleEvent: doesnt connected to a room")
			return
		}

		c.leaveRoom()
	}
}

func (c *client) findRoomByPayload(payload json.RawMessage) *room {
	var roomId uuid.UUID
	err := json.Unmarshal(payload, &roomId)
	if err != nil {
		log.Println("handleEvent: cannot Unmarshal roomId", err)
		return nil
	}
	if r := c.manager.findRoomById(roomId); r != nil {
		return r
	} else {
		log.Println("handleEvent: room not found")
		return nil
	}
}

func (c *client) leaveRoom() {
	if c.currentRoom != nil {
		c.currentRoom.remove <- c
	}
}

func (c *client) joinRoom(r *room) {
	r.add <- c
	c.currentRoom = r
}
