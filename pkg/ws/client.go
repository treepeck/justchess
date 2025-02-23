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
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func newClient(c *websocket.Conn, h *Hub) *client {
	return &client{
		hub:  h,
		conn: c,
		send: make(chan []byte),
	}
}

// readPump pumps and handles all incoming messages from the connection.
func (c *client) readPump(id uuid.UUID) {
	defer func() {
		c.cleanup(id)
	}()

	c.conn.SetReadLimit(512)
	// Set the read deadline for a ping-pong messages to drop inactive connections.
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
			return
		}
		c.handleMsg(id, data)
	}
}

func (c *client) writePump(id uuid.UUID) {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.cleanup(id)
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
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

func (c *client) handleMsg(id uuid.UUID, msg []byte) {
	msgType := msg[len(msg)-1]
	p := make([]byte, 16)
	copy(p[:16], id[:])

	switch msgType {
	case GET_AVAILIBLE_GAMES:
		c.hub.gameEvents <- gameEvent{eType: GET_AVAILIBLE, payload: p}

	case CREATE_GAME:
		p = append(p, []byte{msg[0], msg[1]}...)
		c.hub.gameEvents <- gameEvent{eType: CREATE, payload: p}

	case JOIN_GAME:
		// Append game id.
		p = append(p, msg[:16]...)
		c.hub.gameEvents <- gameEvent{eType: JOIN, payload: p}

	case GET_GAME:
		c.hub.gameEvents <- gameEvent{eType: GET, payload: p}

	case LEAVE_GAME:
		c.hub.gameEvents <- gameEvent{eType: LEAVE, payload: p}
	}
}

// cleanup closes the connection and unregisters the client from the hub.
func (c *client) cleanup(id uuid.UUID) {
	c.conn.Close()
	c.hub.clientEvents <- clientEvent{id: id, sender: nil, eType: UNREGISTER}
}
