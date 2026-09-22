// Package proto defines the communication protocol for game and ws servers.
package proto

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

// Ping is sent by WS server to maintain a TCP connection (keepalive) and measure
// network latency.
type Ping int

// Ping is sent by Game server in response to [Ping].
type Pong int

// Join is sent by WS server to register the player in matchmaking pool.
type Join string

// Leave is sent by WS server to unregister the player from matchmaking pool.
type Leave string

// Create is sent by JustChess to register a new queue or room.
type Create string

// Remove is sent by JustChess to unregister a new queue or room.
type Remove string
