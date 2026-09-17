// Package proto defines the communication protocol for game and ws servers.
package proto

const (
	// Limit of concurrent TCP connections between WS and Game servers.
	MaxConns = 10
)

type Message struct {
	Payload any
}

// Ping is sent by WS server to maintain a TCP connection (keepalive) and measure
// network latency.
type Ping int

// Ping is sent by Game server in response to [Ping].
type Pong int
