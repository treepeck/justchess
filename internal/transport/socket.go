package transport

import (
	"bufio"
	"github.com/treepeck/justchess/pkg/proto"
	"log"
	"net"
)

type socket struct {
	ipc    Ipc
	conn   *net.TCPConn
	reader *bufio.Reader
	writer *bufio.Writer
	pong   chan struct{} // Channel to handle ping messages.
}

func initSocket(ipc Ipc, conn *net.TCPConn) *socket {
	// Wrap connection with buffer to reduce the amount of syscalls.
	// TODO: adjust the buffer size for peformance.
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	s := &socket{
		ipc:    ipc,
		conn:   conn,
		pong:   make(chan struct{}, 5),
		reader: r,
		writer: w,
	}

	go s.read()
	go s.write()

	return s
}

func (s *socket) read() {
	for {
		encoded, err := s.reader.ReadBytes(proto.MessageSeparator)
		if err != nil {
			log.Printf("read error: %v\n", err)
			break
		}
		// Immediately handle ping messages.
		if proto.MessageKind(encoded[0]) == proto.KindPing {
			s.pong <- struct{}{}
		} else {
			m := proto.Decode(encoded)
			// TODO: actually pass the message to queue or room for it to be handled.
			s.ipc.Read <- m
		}
	}

	s.cleanup()
}

func (s *socket) write() {
	for {
		select {
		case encoded := <-s.ipc.Write:
			if _, err := s.writer.Write(encoded); err != nil {
				log.Printf("write error: %v\n", err)
				break
			}
		case <-s.pong:
			if _, err := s.writer.Write(encodedPong); err != nil {
				log.Printf("write error: %v\b", err)
				break
			}
		}
		s.writer.Flush()
	}
}

func (s *socket) cleanup() {
	s.conn.Close()
}
