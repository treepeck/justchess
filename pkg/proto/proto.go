// Package proto defines the communication protocol for game and ws servers.
package proto

const (
	// Limit of concurrent TCP connections between WS and Game servers.
	MaxConns = 10
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
