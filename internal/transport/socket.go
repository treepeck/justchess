package transport

import (
	"bufio"
	// "github.com/treepeck/justchess/pkg/proto"
	"log"
	"net"
)

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
	for {
		// TODO: custom decoder.
		msg, err := s.reader.ReadBytes('\n')
		if err != nil {
			log.Printf("read error: %v\n", err)
			break
		}
		log.Printf("Got a message: %v\n", msg)
	}

	s.cleanup()
}

func (s *socket) write() {
	for {
		msg := <-s.send
		if _, err := s.writer.Write(msg); err != nil {
			log.Printf("write error: %v\n", err)
			break
		}
		s.writer.Flush()
	}
}

func (s *socket) cleanup() {
	s.conn.Close()
}
