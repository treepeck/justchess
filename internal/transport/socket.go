package transport

import (
	"bufio"
	"encoding/gob"
	"github.com/treepeck/justchess/pkg/proto"
	"log"
	"net"
)

type socket struct {
	ipc     Ipc
	conn    *net.TCPConn
	encoder *gob.Encoder
	decoder *gob.Decoder
	reader  *bufio.Reader
	send    chan proto.OutMessage
	writer  *bufio.Writer
}

func initSocket(ipc Ipc, conn *net.TCPConn) *socket {
	// Wrap connection with buffer to reduce the amount of syscalls.
	// TODO: adjust the buffer size for peformance.
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	s := &socket{
		ipc:     ipc,
		conn:    conn,
		encoder: gob.NewEncoder(w),
		decoder: gob.NewDecoder(r),
		send:    make(chan proto.OutMessage, 256),
		reader:  r,
		writer:  w,
	}

	go s.read()
	go s.write()

	return s
}

func (s *socket) read() {
	for {
		var msg proto.InMessage
		if err := s.decoder.Decode(&msg); err != nil {
			log.Printf("decode error: %v\n", err)
			break
		}

		switch t := msg.Payload.(type) {
		case proto.Ping:
			// Immediately respond with pong.
			s.send <- proto.OutMessage{
				Payload: proto.Pong(1),
			}
		// TODO: Route messages sent to game room into ipc.ReadGame.
		case proto.Join, proto.Leave:
			s.ipc.ReadQueue <- msg
		default:
			log.Printf("message has invalid type %v\n", t)
		}
	}

	s.cleanup()
}

func (s *socket) write() {
	for {
		msg := <-s.send
		if err := s.encoder.Encode(msg); err != nil {
			log.Printf("cannot encode message: %v\n", err)
			break
		}
		s.writer.Flush()
	}
}

func (s *socket) cleanup() {
	s.conn.Close()
}
