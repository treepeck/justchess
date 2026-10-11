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
	// Match is used emmited when a new game is being created. The coordinator needs
	// to redirect players and create a new game.
	KindMatch
)

const (
	MessageSeparator byte = 0
	// TODO: maybe payload will need another number of bytes.
	// Encode and Decode functions should use and allocate 0 bytes of memory
	// while dealing with real data. So benchmark that.
	MaxMessageLength = 1 + 12 + 1 + 1 + 1 + 100 + 1 // sysMsgFlag + playerId + sep + messageKind + sep + payload + msgSep.
)

type Message struct {
	Payload json.RawMessage `json:"p"`
	Id      string          `json:"-"`
	Kind    MessageKind     `json:"k"`
}

// Encode encodes [Message] to the binary format.
// NOTE: [KindPing] and [KindPong] messages will not be encoded.
func Encode(m Message) []byte {
	var encoded [MaxMessageLength]byte
	cnt := 1
	encoded[0] = byte(m.Kind)
	switch m.Kind {
	case KindPing, KindPong:
		// NOTE: Skip heartbeat messages.
		return nil
	default:
		for i := range m.Id {
			encoded[cnt] = m.Id[i]
			cnt++
		}
		for i := range m.Payload {
			encoded[cnt] = m.Payload[i]
			cnt++
		}
	}
	encoded[cnt] = MessageSeparator
	return encoded[:cnt+1]
}

// Decode decodes the [Message].
func Decode(encoded []byte) Message {
	return Message{
		Kind:    MessageKind(encoded[0]),
		Payload: encoded[13 : len(encoded)-1],
		Id:      string(encoded[1:13]),
	}
}
