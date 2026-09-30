package transport

import (
	"bufio"
	"github.com/treepeck/justchess/pkg/proto"
	"log"
	"net"
)

var encodedPong = []byte{byte(proto.KindPong), byte(proto.MessageSeparator)}

type socket struct {
	ipc    Ipc
	conn   *net.TCPConn
	reader *bufio.Reader
	writer *bufio.Writer
	send   chan []byte
}

func initSocket(ipc Ipc, conn *net.TCPConn) *socket {
	// Wrap connection with buffer to reduce the amount of syscalls.
	// TODO: adjust the buffer size for peformance.
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	s := &socket{
		ipc:    ipc,
		conn:   conn,
		send:   make(chan []byte, 256),
		reader: r,
		writer: w,
	}

	go s.read()
	go s.write()

	return s
}

func (s *socket) read() {
	parts := proto.PreallocateDecodeBuff()

	// TODO: might be a race condition. Need to test that.
	encoded := make([]byte, proto.MaxMessageLength)
	var m proto.Message
	var err error
	for {
		encoded, err = s.reader.ReadBytes(proto.MessageSeparator)
		if err != nil {
			log.Printf("read error: %v\n", err)
			break
		}
		// Immediately handle ping messages.
		if proto.MessageKind(encoded[0]) == proto.KindPing {
			s.send <- encodedPong
		} else {
			m = proto.Decode(parts, encoded, proto.MessageKind(encoded[0]))
			log.Printf("Got a message: %v\n", m)
		}
	}

	s.cleanup()
}

func (s *socket) write() {
	for {
		encoded := <-s.send
		if _, err := s.writer.Write(encoded); err != nil {
			log.Printf("write error: %v\n", err)
			break
		}
		s.writer.Flush()
	}
}

func (s *socket) cleanup() {
	s.conn.Close()
}
