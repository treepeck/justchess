// Package proto defines the communication protocol for game and ws servers.
package proto

import (
	"encoding/gob"
)

const (
	// Limit of concurrent TCP connections between WS and Game servers.
	MaxConns = 10
	// Limit of concurrent clients over a single TCP connection. If the number
	// exceeds the limit, a new connection must be opened.
	ClientsPerConn = 500
)

// InMessage is a message sent by WebSocket server.
type InMessage struct {
	PlayerId string
	Payload  any
}

// OutMessage is a message sent by JustChess server.
type OutMessage struct {
	Recievers []string
	Payload   any
}

// Ping [InMesage] used to maintain a TCP connection (keepalive) and measure network latency.
type Ping int

// Pong is [OutMessage] sent in response to [Ping].
type Pong int

// Join [InMessage] used to register a player in matchmaking pool.
type Join string

// Leave is [InMessage] used to unregister a player from matchmaking pool.
type Leave string

// Counter is [OutMessage] used to notify a player about number of other players in matchmaking.
type Counter int

// Redirect is [OutMessage] used to redirect a player to named URL.
type Redirect string

func RegisterGOBTypes() {
	gob.Register(Ping(0))
	gob.Register(Pong(0))
	gob.Register(Join(""))
	gob.Register(Leave(""))
	gob.Register(Counter(0))
	gob.Register(Redirect(""))
}
