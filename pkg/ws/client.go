package ws

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type color int

const (
	// If the client is playing with white pieces, his color will be white.
	colorWhite color = iota
	// If the client is playing with black pieces, his color will be black.
	colorBlack
	// In case the client is just spectator.
	colorNone
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "http://localhost:3502"
	},
}

func HandleNewConnection(room *Room, rw http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Printf("Cannot upgrade connection %v with IP: %s", err, r.RemoteAddr)
		return
	}

	client := NewClient(room, conn)

	log.Printf("New connection %s", client.id)

	go client.readRoutine()
	go client.writeRoutine()

	// Register client in the room.
	room.gate <- client
}

type client struct {
	// id is random string to differentiate clients.
	id         string
	connection *websocket.Conn
	// send channel must be buffered, otherwise if the routine writes to it but the client
	// drops connection, the routine will wait forever.
	send chan []byte
	// To be able to send messages to the room.
	room *Room
	// Is WebSocket connection active.
	isConnected bool
	// Whether the client have created the room it connected to.
	isRoomCreator bool
	color         color
}

func NewClient(r *Room, conn *websocket.Conn) *client {
	return &client{
		id:            rand.Text(),
		connection:    conn,
		send:          make(chan []byte, 256),
		room:          r,
		isConnected:   true,
		isRoomCreator: false,
		color:         colorNone,
	}
}

func (c *client) readRoutine() {
	defer func() {
		c.cleanup()
	}()

	c.connection.SetReadLimit(maxMessageSize)
	c.connection.SetReadDeadline(time.Now().Add(pongWait))
	c.connection.SetPongHandler(c.pongHandler)

	for {
		_, rawMsg, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected disconnection: %v ID: %s", err, c.id)
			}
			return
		}

		c.handleMessage(rawMsg)
	}
}

func (c *client) writeRoutine() {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		c.cleanup()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
			log.Printf("JSON message to %s", c.id)

		// Send ping messages periodically.
		case <-pingTicker.C:
			c.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			log.Printf("Ping message to %s", c.id)
		}
	}
}

func (c *client) handleMessage(rawMsg []byte) {
	var msg message
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		log.Printf("Recieved invalid message from %s %v", c.id, err)
		return
	}

	switch msg.Action {
	case actionMakeMove:
		if c.room == nil {
			log.Printf("Recieved move message but client is not in the room %s", c.id)
			return
		}

		var p makeMovePayload
		err := json.Unmarshal([]byte(msg.Payload), &p)
		if err != nil {
			log.Printf("Cannot Unmarshal move info: %v %s", err, c.id)
			return
		}
		p.senderId = c.id
		c.room.move <- p

	default:
		log.Printf("Recieved message with invalid action %s", c.id)
	}
}

func (c *client) cleanup() {
	if c.isConnected {
		c.connection.Close()
		log.Printf("Disconnected %s", c.id)
		c.isConnected = false

		// Unregister client from the room.
		c.room.gate <- c
	}
}

func (c *client) pongHandler(appData string) error {
	if len(appData) > 0 {
		log.Printf("WARNING: non-empty ping message from %s", c.id)
	}

	log.Printf("Pong message from %s", c.id)
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}
