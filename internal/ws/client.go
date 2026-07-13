package ws

import (
	"github.com/gorilla/websocket"
	"time"
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

// client is a wrapper around a single WebSocket connection.
type client struct {
	// Timestamp when the last [messagePing] was sent.
	pingTimestamp time.Time
	conn          *websocket.Conn
	// send is a channel which recieves messages that the client will write to
	// the WebSocket connection.  It must recieve raw bytes to avoid expensive
	// encoding for each client in case of message broadcasting. The reason for
	// the unbuffered send channel is that Gorilla WebSocket library allows only
	// one concurrent writer to a connection at a time.
	send chan []byte
	// pong is used to prevent race conditions while handling [messagePong].
	pong chan struct{}
	// ping is is a network latency in milliseconds. To calculate it, a full
	// roundtrip is performed.
	ping int
	// New [messagePing] must be sent only when the client doesn't have a
	// pending one. Otherwise the delay cannot be correctly measured.
	hasPendingPing bool
}

// initClient returns [client] with configured connection parameters.
// also runs client's goroutines.
func initClient(conn *websocket.Conn) *client {
	c := &client{
		conn: conn,
		send: make(chan []byte, 192),
		pong: make(chan struct{}, 10),
	}

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))

	go c.read()
	go c.write()

	return c
}

// read sequentially reads messages from the connection.
//
// If the message cannot be properly read or decoded, the connection is
// instantly closed.
func (c *client) read() {
	for {
		var m message
		if err := c.conn.ReadJSON(&m); err != nil {
			return
		}

		if m.Kind == messagePong {
			c.pong <- struct{}{}
		} else {
			m.SenderId = c.id
		}
	}
}

// write sequentially takes the incomming messages from the send channel and
// writes them to the connection. It also sends [pingMessage]s each
// [pingPeriod] seconds to maintain a heartbeat.
func (c *client) write() {
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	for {
		select {
		case raw, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, raw); err != nil {
				continue
			}
		// heartbeat is a mechanism that is necessary for several reasons:
		//   - Detect inactive connections and free resources after a player leaves the page.
		//   - Measure ping to provide a fairer gameplay experience.
		//   - Prevent the browser from automatically closing an idle client-side connection
		//     after a few minutes without outgoing messages. A player may spend several
		//     minutes thinking about a move, so periodically sending a heartbeat in the
		//     background is necessary to keep the connection alive.
		case <-c.pong:
			// Handle pong messages only when the client has a pending ping.
			// Otherwise the latency cannot be correctly measured. Important
			// note is WebSocket protocol doesn't allow dropped frames, meaning
			// all messages are eventually delivered.
			if c.hasPendingPing {
				c.hasPendingPing = false
				c.ping = int(time.Since(c.pingTimestamp).Milliseconds())
				if c.conn.SetReadDeadline(time.Now().Add(pongWait)) != nil {
					return
				}
			}
		case <-pingTicker.C:
			// Send a new [messagePing] only if the previous one is answered.
			if c.hasPendingPing {
				continue
			}

			c.pingTimestamp = time.Now()
			c.conn.SetWriteDeadline(c.pingTimestamp.Add(writeWait))

			if err := c.conn.WriteJSON(message{
				Kind:    messagePing,
				Payload: []byte(strconv.Itoa(c.ping)),
			}); err != nil {
				continue
			}
			c.hasPendingPing = true
		}
	}
}
