package transport

import (
	"encoding/gob"
	"errors"
	"github.com/treepeck/justchess/pkg/proto"
	"log"
	"net"
	"os"
	"strconv"
	"sync/atomic"
)

const maxTopics = 109 // 100 games and 9 queues.

// Service communicates with WebSocket server via a dynamic pool of TCP connections.
//
// TODO: reconnection.
// TODO: acknowledge.
// TODO: The server should not run if it cannot connect to WS server.
type Service struct {
	listener    *net.TCPListener
	in          chan *net.TCPConn
	out         chan *socket
	isListening *atomic.Bool
	// Set of active TCP sockets.
	sockets map[*socket]struct{}
	topics  map[string]topic
}

// InitService initializes the [Service] and listens the TCP network.
// NOTE: Connections are blindly accepted. The TCP port should be
// guarded by the OS firewall.
func InitService() (Service, error) {
	s := Service{
		in:          make(chan *net.TCPConn),
		out:         make(chan *socket),
		isListening: &atomic.Bool{},
		sockets:     make(map[*socket]struct{}, proto.MaxConns),
		topics:      make(map[string]topic, maxTopics),
	}
	s.isListening.Store(true)

	// TODO: maybe extract it to some other place.
	gob.Register(proto.Ping(0))
	gob.Register(proto.Pong(0))

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

	go s.listen()
	go s.eventBus()

	return s, nil
}

// listen listens for TCP connections. Listen only works when the amount of active
// connections doesn't reach [maxConns].
func (s Service) listen() {
	for {
		if !s.isListening.Load() {
			break
		}

		c, err := s.listener.AcceptTCP()
		if err != nil {
			log.Printf("cannot accept connection: %v\n", err)
		}
		s.in <- c
	}
}

func (s Service) eventBus() {
	defer s.cleanup()

	for {
		select {
		case conn := <-s.in:
			s.register(conn)
		case sock := <-s.out:
			s.unregister(sock)
		}
	}
}

func (s Service) register(conn *net.TCPConn) {
	if len(s.sockets) == proto.MaxConns {
		log.Printf("limit of TCP connections was exceeded\n")
		// Stop listening for TCP connections until the limit is statisfied.
		s.isListening.Store(false)
		// Close the incomming connection as it cannot be maintained.
		conn.Close()
		return
	}

	sock := initSocket(conn)
	s.sockets[sock] = struct{}{}
	log.Printf("open %v conn\n", conn)
}

func (s Service) unregister(sock *socket) {
	if _, ok := s.sockets[sock]; !ok {
		return
	}

	delete(s.sockets, sock)
	if len(s.sockets) < proto.MaxConns && !s.isListening.Load() {
		// Continue listening for TCP connections.
		s.isListening.Store(true)
		go s.listen()
	}
	log.Printf("close %v conn\n", sock)
}

// cleanup is called only in case the server crushes. It is needed to gracefully
// terminate TCP connections without losing packets.
func (s Service) cleanup() {
	if err := s.listener.Close(); err != nil {
		log.Printf("cannot close TCP listener: %v\n", err)
	}
	// TODO: close all remaining connections.
}
