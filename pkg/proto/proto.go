// Package proto defines the communication protocol for game and ws servers.
package proto

import "encoding/json"

const (
	// Limit of concurrent TCP connections between WS and Game servers.
	MaxConns = 10
	// Limit of concurrent clients over a single TCP connection. If the number
	// exceeds the limit, a new connection must be opened.
	ClientsPerConn = 500
)

type MessageKind int

const (
	// Ping used to maintain a TCP connection (keepalive) and measure network latency.
	// Payload is an integer that represents the latency in milliseconds.
	KindPing MessageKind = iota
	// Pong is sent in response to [Ping]. Payload is nil.
	KindPong
	// Join used to register a player in matchmaking pool. Payload is nil.
	KindJoin
	// Leave is used to unregister a player from matchmaking pool. Payload is nil.
	KindLeave
	// Counter is used to notify a player about number of players in matchmaking queue.
	// Payload is an integer.
	KindCounter
	// Redirect is used to redirect a player to named URL.
	KindRedirect
)

type Message struct {
	Kind     MessageKind     `json:"k"`
	Payload  json.RawMessage `json:"p"`
}
