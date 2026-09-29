// Package proto defines the communication protocol for game and ws servers.
package proto

import (
	"bytes"
	"errors"
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

const (
	messagePartSeparator byte = ' '
	// TODO: maybe payload will need another number of bytes.
	// Encode and Decode functions should use and allocate 0 bytes of memory
	// while dealing with real data. So benchmark that.
	maxMessageLength = 12 + 1 + 100 // playerId + messageKind + payload.
)

var errNotValid = errors.New("proto: message is not valid")

type Message struct {
	Payload []byte
	Id      string
	Kind    MessageKind
}

// Encoder wraps a bytes.Buffer and implements the encoding part of the protocol.
type Encoder struct {
	buff bytes.Buffer
}

func NewEncoder() *Encoder {
	var b bytes.Buffer
	b.Grow(maxMessageLength)

	return &Encoder{
		buff: b,
	}
}

func (e *Encoder) Encode(m Message) []byte {
	e.buff.Reset()
	// We do not need to check errors returned by Buffer methods, because they are
	// always nil.
	e.buff.WriteString(m.Id)
	e.buff.WriteByte(messagePartSeparator)
	e.buff.WriteByte(byte(m.Kind))
	e.buff.WriteByte(messagePartSeparator)
	e.buff.Write(m.Payload)
	return e.buff.Bytes()
}

func Decode(parts [][]byte, encoded []byte) (Message, error) {
	curr := 0
	j := 0
	for i := range encoded {
		if encoded[i] == messagePartSeparator {
			curr++
			j = 0
		} else {
			parts[curr][j] = encoded[i]
			j++
		}
	}

	// Validate the message.
	if len(parts[0]) != /* ID len */ 12 || len(parts[1]) != 1 {
		return Message{}, errNotValid
	}

	return Message{
		Id:      string(parts[0]),
		Kind:    MessageKind(parts[1][0]),
		Payload: parts[2],
	}, nil
}
