package ws

import (
	"justchess/pkg/auth"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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

// client wraps a single active connection and provides methods for reading,
// writing WebSocket and handling WebSocket messages.
type client struct {
	// id must be equal to player_id in the database.
	id int64
	// The id of the topic (hub or room) which outcomming events the client will recieve.
	// An empty string means that client is subscribed to the Hub's events.
	subscribtionId string
	connection     *websocket.Conn
	// send channel must be buffered, otherwise if the goroutine writes to it but the client
	// drops the connection, the goroutine will wait forever.
	send chan event
	// To be able to send events to the hub.
	hub *Hub
	// Is WebSocket connection alive.
	isConnected bool
}

// HandleNewConnection creates a new client and runs [client.read] and [client.write]
// methods as goroutines. Each handshake request must include a context with a valid playerId
// to identify the client. If the handshake request includes a roomId query parameter,
// the hub will check if the room exists and connect the client to it.
func HandleNewConnection(h *Hub, rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context().Value(auth.PidKey)
	if ctx == nil {
		log.Printf("ERROR: handshake request with nil context")
		http.Error(rw, "Missing player id", http.StatusUnauthorized)
		return
	}
	pid := ctx.(int64)

	// If rid is missing or not valid, the client will be subscribed to hub.
	rid := r.URL.Query().Get("roomId")

	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Printf("Cannot upgrade connection %v with IP: %s", err, r.RemoteAddr)
		return
	}

	c := &client{
		id:             pid,
		hub:            h,
		connection:     conn,
		send:           make(chan event, 256),
		isConnected:    true,
		subscribtionId: rid,
	}

	go c.read()
	go c.write()

	h.register <- c
}

// read consequentially (one at a time) reads messages from the connection.
// Handles pong messages to maintain a heartbeat.
// It is designed to run as a goroutine while the connection is alive.
func (c *client) read() {
	defer func() {
		c.cleanup()
	}()

	c.connection.SetReadLimit(maxMessageSize)
	c.connection.SetReadDeadline(time.Now().Add(pongWait))
	c.connection.SetPongHandler(c.pongHandler)

	for {
		var e event
		err := c.connection.ReadJSON(&e)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Printf("Client %d unexpected disconnection: %v", c.id, err)
			}
			return
		}

		e.PubId = c.id
		e.TopicId = c.subscribtionId

		c.hub.bus <- e
	}
}

// write consequentially (one at a time) writes messages to the connection.
// Automatically sends ping messages to maintain heartbeat.
// Designed to run as a goroutine while the connection is alive.
func (c *client) write() {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		c.cleanup()
	}()

	for {
		select {
		case e, ok := <-c.send:
			c.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.connection.WriteJSON(e); err != nil {
				return
			}

		// Send ping messages periodically.
		case <-pingTicker.C:
			c.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// cleanup closes the connection and unregisters the client from the Hub.
func (c *client) cleanup() {
	if c.isConnected {
		c.isConnected = false
		c.connection.Close()
		c.hub.unregister <- c
	}
}

// pongHandler handles the incomming pong messages to maintain a heartbeat.
// Heartbeat helps to drop inactive while maintaining the active connections.
func (c *client) pongHandler(appData string) error {
	if len(appData) > 0 {
		log.Printf("WARNING: Client %d send non-empty pong message", c.id)
	}
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}
