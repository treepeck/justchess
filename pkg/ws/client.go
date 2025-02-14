package ws

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Deadline for the next pong message from peer.
	pongWait = 10 * time.Second
	// Client sends ping messages with the defined interval.
	// It must be less than pongWait.
	pingInterval = (pongWait * 9) / 10
)

type client struct {
	id          uuid.UUID
	conn        *websocket.Conn
	manager     *Manager
	currentRoom *room
	send        chan []byte
}

func newClient(c *websocket.Conn, m *Manager) *client {
	return &client{
		id:          uuid.New(),
		conn:        c,
		manager:     m,
		currentRoom: nil,
		send:        make(chan []byte),
	}
}

// readPump pumps and handles all incoming messages from the connection.
func (c *client) readPump() {
	defer func() {
		c.conn.Close()
		if c.currentRoom != nil {
			c.currentRoom.unregister <- c
			c.currentRoom = nil
		}
		c.manager.unregister <- c
	}()

	c.conn.SetReadLimit(512)
	// Set the read deadline to limit inactive connections.
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("%v\n", err)
		return
	}
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		msgType, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway,
				websocket.CloseNormalClosure) {
				log.Printf("%v\n", err)
			}
			return
		}

		if msgType != websocket.BinaryMessage {
			continue
		}
		c.handleMsg(data)
	}
}

func (c *client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		if c.currentRoom != nil {
			c.currentRoom.unregister <- c
			c.currentRoom = nil
		}
		c.manager.unregister <- c
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				log.Printf("%v\n", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("%v\n", err)
				return
			}
		}
	}
}

func (c *client) handleMsg(msg []byte) {
	if len(msg) < 1 {
		return
	}

	msgType := msg[len(msg)-1]
	switch msgType {
	case CREATE_ROOM:
		// Forbit multiple room creation at a time.
		if c.currentRoom != nil {
			return
		}
		r := newRoom(msg[0], msg[1])
		go r.run()
		c.manager.add <- r
		r.register <- c

	case JOIN_ROOM:
		// Forbit multiple room joining at a time.
		if c.currentRoom != nil {
			return
		}

		id, err := uuid.FromBytes(msg[0:16])
		if err != nil { // Invalid room id.
			log.Printf("%v\n", err)
			return
		}

		for r := range c.manager.rooms {
			if r.id == id {
				r.register <- c
			}
		}

	case LEAVE_ROOM:
		if c.currentRoom == nil {
			return
		}
		c.currentRoom.unregister <- c
		c.currentRoom = nil

	case MOVE:
		if c.currentRoom == nil {
			return
		}
		c.currentRoom.moves <- msg
	}
}
