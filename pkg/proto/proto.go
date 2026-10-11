// Package proto defines the communication protocol for game and ws servers.
package proto

import (
	"encoding/json"
	"github.com/treepeck/justchess/internal/randgen"
)

const (
	// Limit of concurrent TCP connections between WS and Game servers.
	MaxConns = 10
	// Limit of concurrent clients over a single TCP connection. If the number
	// exceeds the limit, a new connection must be opened.
	ClientsPerConn = 500
)

type MessageKind int

const (
	// KindPing used to maintain a TCP connection (keepalive) and measure network latency.
	// Payload is a byte sequence (1 or more bytes) that represents the latency in milliseconds.
	KindPing MessageKind = iota
	// KindPong is sent in response to [Ping]. Payload is nil.
	KindPong
	// KindJoin used to register a player in matchmaking pool. Payload is nil.
	KindJoin
	// KindLeave is used to unregister a player from matchmaking pool. Payload is nil.
	KindLeave
	// KindCounter is used to notify a player about number of players in matchmaking queue.
	// Payload is an integer encoded in a byte sequence.
	KindCounter
	// KindAdd is used to notify Coordinator about new topic. Payload is nil.
	KindAdd
	// KindRemove is used to notify Coordinator about removed topic. Payload is nil.
	KindRemove
	// KindMatch is emmited when a new game is being created.
	// Payload is [Player1ID_Player2ID_GameId]. Note that '_' is used only for readability
	// and is ommited in real message.
	KindMatch
)

const (
	MessageSeparator byte = 255
	// TODO: maybe payload will need another number of bytes.
	// Encode and Decode functions should use and allocate 0 bytes of memory
	// while dealing with real data. So benchmark that.
	MaxMessageLength = 1 + 12 + 100 + 1 // MessageKind + Id + payload + MessageSeparator.
)

type Message struct {
	Payload json.RawMessage `json:"p"`
	Id      string          `json:"-"`
	Kind    MessageKind     `json:"k"`
}

// Encode encodes [Message] to the binary format. It's the caller's responsibility
// to encode a payload since it might be a JSON or an arbitrary array of bytes. For
// definition of payloads see the documentation of specific [MessageKind]'s.
// NOTE: [KindPing] and [KindPong] messages will not be encoded.
func Encode(m Message) []byte {
	var encoded [MaxMessageLength]byte
	encoded[0] = byte(m.Kind)
	switch m.Kind {
	case KindPing, KindPong:
		// NOTE: Skip heartbeat messages.
		return nil
	default:
		copy(encoded[1:randgen.IdLen+1], m.Id)
		copy(encoded[randgen.IdLen+1:], m.Payload)
	}
	end := randgen.IdLen + len(m.Payload) + 1
	encoded[end] = MessageSeparator
	return encoded[:end+1]
}

// Decode decodes the [Message].
func Decode(encoded []byte) Message {
	return Message{
		Kind:    MessageKind(encoded[0]),
		Payload: encoded[randgen.IdLen+1 : len(encoded)-1],
		Id:      string(encoded[1 : randgen.IdLen+1]),
	}
}
