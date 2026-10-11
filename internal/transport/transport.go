package transport

import (
	"errors"
	// "github.com/treepeck/justchess/internal/game"
	"github.com/treepeck/justchess/internal/matchmaking"
	"github.com/treepeck/justchess/pkg/proto"
	"log"
	"net"
	"os"
	"strconv"
	"sync/atomic"
)

const maxHandlers = 109 // 100 games and 9 queues.

// Ipc wraps all channels used for inter-process communication between [transport]
// and [game] packages.
type Ipc struct {
	Write chan []byte
	Read  chan proto.Message
}

// Service manages the dynamic pool of TCP connections with the Coordinator server.
type Service struct {
	Ipc         Ipc
	listener    *net.TCPListener
	open        chan *net.TCPConn
	close       chan *socket
	isListening *atomic.Bool
	// Set of active TCP sockets.
	sockets map[*socket]struct{}
	// handlers maps the handler ID to the handlers's channel.
	// A handler is a game or matchmaking queue.
	handlers map[string]chan proto.Message
}

// InitService initializes the [Service] and listens the TCP network.
// NOTE: Connections are blindly accepted. The TCP port should be
// guarded by the OS firewall.
func InitService() (Service, error) {
	s := Service{
		open:  make(chan *net.TCPConn),
		close: make(chan *socket),
		Ipc: Ipc{
			Write: make(chan []byte, 256),
			Read:  make(chan proto.Message, 256),
		},
		isListening: &atomic.Bool{},
		handlers:    make(map[string]chan proto.Message, maxHandlers),
		sockets:     make(map[*socket]struct{}, proto.MaxConns),
	}

	// Manually add a handler for each matchmaking queue.
	for i := range 9 {
		id := string(i + '0')
		s.handlers[id] = matchmaking.SpawnQueue(id, s.Ipc.Write)
	}

	addr := net.ParseIP(os.Getenv("JUSTCHESS_TCP_ADDR"))
	if addr == nil {
		return Service{}, errors.New("game: JUSTCHESS_TCP_ADDR environment variable must be not nil")
	}

	port, err := strconv.Atoi(os.Getenv("JUSTCHESS_TCP_PORT"))
	if err != nil {
		return Service{}, err
	}

	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   addr,
		Port: port,
	})
	if err != nil {
		return Service{}, err
	}
	s.listener = l

	go s.accept()
	go s.listen()

	return s, nil
}

// accept listens for TCP connections. It should only run when the amount of active
// connections doesn't reach [maxConns].
func (s Service) accept() {
	s.isListening.Store(true)

	for {
		if !s.isListening.Load() {
			break
		}

		c, err := s.listener.AcceptTCP()
		if err != nil {
			log.Printf("cannot accept connection: %v\n", err)
		}
		s.open <- c
	}
}

func (s Service) listen() {
	defer s.cleanup()

	for {
		select {
		case conn := <-s.open:
			s.openSocket(conn)
		case sock := <-s.close:
			s.closeSocket(sock)
		case m := <-s.Ipc.Read:
			s.route(m)
		}
	}
}

func (s Service) openSocket(conn *net.TCPConn) {
	if len(s.sockets) == proto.MaxConns {
		log.Print("TCP connection limit reached")
		// Stop listening for TCP connections until the limit is statisfied.
		s.isListening.Store(false)
		// Close the incomming connection as it cannot be maintained.
		conn.Close()
		return
	}

	sock := initSocket(s.Ipc, conn)
	s.sockets[sock] = struct{}{}
	log.Printf("opened new TCP socket %v\n", sock)
}

// closeSocket closes a socket.
func (s Service) closeSocket(sock *socket) {
	if _, ok := s.sockets[sock]; !ok {
		return
	}

	delete(s.sockets, sock)
	sock.conn.Close() // TODO: might want to handle error.
	if len(s.sockets) < proto.MaxConns && !s.isListening.Load() {
		// Continue listening for TCP connections.
		go s.accept()
	}
	log.Printf("closed TCP socket %v\n", sock)
}

func (s Service) route(m proto.Message) {
	switch m.Kind {
	// For those messages, the payload contains the handler URL.
	// So the message must be simply forwarded to the named handler.
	case proto.KindJoin, proto.KindLeave:
		in, ok := s.handlers[string(m.Payload)]
		if !ok {
			log.Printf("player %s sends message to non-existing handler %s\n", m.Id, m.Payload)
			return
		}
		in <- m
	}
}

// cleanup is called only in case the server crushes. It is needed to gracefully
// terminate TCP connections without losing packets.
func (s Service) cleanup() {
	if err := s.listener.Close(); err != nil {
		log.Printf("cannot close TCP listener: %v\n", err)
	}
	// TODO: close all remaining connections.
}
