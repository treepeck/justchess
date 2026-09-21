package transport

import (
	"bufio"
	"encoding/gob"
	"github.com/treepeck/justchess/pkg/proto"
	"log"
	"net"
)

type socket struct {
	conn    *net.TCPConn
	encoder *gob.Encoder
	decoder *gob.Decoder
	reader  *bufio.Reader
	writer  *bufio.Writer
}

func initSocket(conn *net.TCPConn) *socket {
	// Wrap connection with buffer to reduce the amount of syscalls.
	// TODO: adjust the buffer size for peformance.
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	s := &socket{
		conn:    conn,
		encoder: gob.NewEncoder(w),
		decoder: gob.NewDecoder(r),
		reader:  r,
		writer:  w,
	}

	go s.listen()

	return s
}

func (s *socket) listen() {
	defer s.cleanup()

	for {
		var msg proto.Message
		if err := s.decoder.Decode(&msg); err != nil {
			log.Printf("decode error: %v\n", err)
			break
		}

		switch t := msg.Payload.(type) {
		case proto.Ping:
			// Immediately respond with pong.
			s.encoder.Encode(proto.Message{
				Payload: proto.Pong(1),
			})
			s.writer.Flush()
			log.Printf("pong")
		case proto.Move:

		default:
			log.Printf("message has invalid type: %v\n", t)
		}
	}
}

func (s *socket) cleanup() {
	s.conn.Close()
}
