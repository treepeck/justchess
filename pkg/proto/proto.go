// Package proto defines the communication protocol for game and ws servers.
package proto

import (
	"github.com/treepeck/justchess/internal/randgen"
)

const (
	// Limit of concurrent TCP connections between WS and Game servers.
	MaxConns = 10
	// Limit of concurrent clients over a single TCP connection. If the number
	// exceeds the limit, a new connection must be opened.
	ClientsPerConn = 500
)

type MessageKind = byte

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

// 0: [MessageKind];
// 1-12: Player or topic ID;
// 13-MessageSeparator: Optional payload.
type Message struct {
	Payload []byte // Arbitrary encoded payload which concrete type depends on message kind.
	Id      string
	Kind    MessageKind
}

// Encode does NOT encode the payload of the [Message]. It's the caller's responsibility
// to pass the encoded payload. The goal of Encode function is to be inlined and therefore
// callers must call additional functions which are also inlined to encode the full message.
func Encode(m Message) []byte {
	// Array of bytes on stack.
	var encoded [MaxMessageLength]byte
	encoded[0] = m.Kind
	copy(encoded[1:randgen.IdLen+1], m.Id)
	copy(encoded[randgen.IdLen+1:], m.Payload)
	encoded[randgen.IdLen+1+len(m.Payload)] = MessageSeparator
	return encoded[:randgen.IdLen+1+len(m.Payload)+1]
}

// EncodeInt splits integer into the byte sequence and returns it.
// NOTE: The resulting sequence will never have [MessageSeparator] byte in it.
// Each byte of the x should be encoded with a continuation bit because a single
// payload may contain more than one number.
func EncodeInt(x uint16) []byte {
	var seq [3]byte
	// Write 6 bits at a time.
	i := 0
	for ; ; i++ {
		chunk := byte(x & 0x3F)
		seq[i] = chunk
		x >>= 6
		if x == 0 {
			break
		}
		seq[i] |= 0x40 // Set continuation bit.
	}
	return seq[:i+1]
}

// DecodeInt restores the interger from the encoded byte sequence.
// The key in decoding is to understand that each byte in the sequence
// must be smaller than previous one. Otherwise it belongs to the next number.
func DecodeInt(encoded []byte) uint16 {
	var x, coeff uint16
	for i := range encoded {
		x |= uint16(encoded[i]&0x3F) << coeff
		coeff += 6
	}
	return x
}

// Decode decodes the [Message].
func Decode(encoded []byte) Message {
	return Message{
		Kind:    MessageKind(encoded[0]),
		Payload: encoded[randgen.IdLen+1 : len(encoded)-1],
		Id:      string(encoded[1 : randgen.IdLen+1]),
	}
}
