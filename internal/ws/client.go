package ws

import (
	"log"
	"strconv"
	"time"

	"justchess/internal/db"

	"github.com/gorilla/websocket"
)

// Connection parameters.
const (
	//  Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 7 * time.Second
	// Send pings to peer with this period.  Must be less than pongWait.
	pingPeriod = 3 * time.Second
	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

// client wraps a single socket and player info.
type client struct {
	player db.Player
	// Timestamp when the last ping event was sent to measure the network delay.
	pingTimestamp time.Time
	// send is a channel which recieves events that the client will write to
	// the WebSocket connection.  It must recieve raw bytes to avoid expensive
	// JSON encoding for each client in case of event broadcasting.
	// The reason for the send channel is that events must be read and written
	// sequentially, since the Gorilla WebSocket library allows only one
	// concurrent writer to a connection at a time.
	send chan []byte
	// forward is a channel to which the client will send events.
	forward chan event
	// unregister is a channel to which the client will send to unregister themself
	// from the room or queue.
	unregister chan string
	conn       *websocket.Conn
	// Network delay in milliseconds.
	ping int
	// New ping event must be sent only when the client responses to the
	// previous one.  Otherwise the delay cannot be correctly measured.
	hasAnsweredPing bool
}

// newClient creates a new client and sets the connection properties.
func newClient(conn *websocket.Conn, p db.Player) *client {
	now := time.Now()

	c := &client{
		player:        p,
		send:          make(chan []byte, 192),
		conn:          conn,
		ping:          0,
		pingTimestamp: now,
		// Must be true to be able to send the first ping message.
		hasAnsweredPing: true,
	}

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(now.Add(pongWait))

	return c
}

// read reads and handles events from the connection sequentially (one at a time).
//
// Pong events are handled by the client itself.  In the case of other event,
// they are forwarded to the forward channel.  If an event cannot be read, the
// connection will be closed.
func (c *client) read() {
	defer c.cleanup()

	var e event
	for {
		if err := c.conn.ReadJSON(&e); err != nil {
			return
		}

		if e.Action == actionPong {
			c.handlePong()
		} else {
			if c.forward != nil {
				e.sender = c
				c.forward <- e
			}
		}
	}
}

// write takes the incomming events from the send channel and writes them to the
// connection sequentially (one at a time).
//
// Automatically sends ping events to maintain a heartbeat.
func (c *client) write() {
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	for {
		select {
		case raw, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, raw); err != nil {
				return
			}

		// Send ping messages periodically.
		case <-pingTicker.C:
			// Send a new ping event only if the client has already answered to
			// the previous one.
			if !c.hasAnsweredPing {
				continue
			}

			now := time.Now()
			c.conn.SetWriteDeadline(now.Add(writeWait))

			c.pingTimestamp = now

			if err := c.conn.WriteJSON(event{
				Action:  actionPing,
				Payload: []byte(strconv.Itoa(c.ping)),
			}); err != nil {
				log.Println(err)
				return
			}
			c.hasAnsweredPing = false
		}
	}
}

// handlePong handles the incomming pong messages to maintain a heartbeat.
//
// Sending ping and pong messages is necessary because without it the connections
// are interrupted after about 2 minutes of no message sending from the client.
//
// Sets the delay value to the time elapsed since the last ping was sent.  This
// helps determine an up-to-date network delay value, which will be subtracted from
// the player's clock to provide a fairer gameplay experience.
func (c *client) handlePong() error {
	// Handle pong events only when the client has a pending ping event.
	if !c.hasAnsweredPing {
		c.hasAnsweredPing = true
		c.ping = int(time.Since(c.pingTimestamp).Milliseconds())
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	}
	return nil
}

// cleanup closes the connection and unregisters the client from the gatekeeper.
func (c *client) cleanup() {
	if c.unregister != nil {
		c.unregister <- c.player.Id
	}
	close(c.send)
	c.conn.Close()
}
