package transport

import (
	"errors"
	// "github.com/treepeck/justchess/internal/game"
	"github.com/treepeck/justchess/pkg/proto"
	"log"
	"net"
	"os"
	"strconv"
	"sync/atomic"
)

const maxTopics = 109 // 100 games and 9 queues.

// Ipc wraps all channels used for inter-process communication between [transport]
// and [game] packages.
type Ipc struct {
	Write  chan []byte
	Read   chan []byte
	Add    chan Topic
	Remove chan string // To remove [Topic] by ID.
}

type Topic struct {
	Id string
	In chan proto.Message
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
	// topics maps the handler ID to the topics's channel.
	// A handler is a game or matchmaking queue.
	topics map[string]chan proto.Message
}

// InitService initializes the [Service] and listens the TCP network.
// NOTE: Connections are blindly accepted. The TCP port should be
// guarded by the OS firewall.
func InitService() (Service, error) {
	s := Service{
		open:  make(chan *net.TCPConn),
		close: make(chan *socket),
		Ipc: Ipc{
			Write:  make(chan []byte, 256),
			Read:   make(chan []byte, 256),
			Add:    make(chan Topic),
			Remove: make(chan string),
		},
		isListening: &atomic.Bool{},
		topics:      make(map[string]chan proto.Message, maxTopics),
		sockets:     make(map[*socket]struct{}, proto.MaxConns),
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
		case t := <-s.Ipc.Add:
			s.addTopic(t)
		case topicId := <-s.Ipc.Remove:
			s.removeTopic(topicId)
		case encoded := <-s.Ipc.Read:
			s.route(encoded)
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
	log.Printf("opened new TCP socket %p\n", sock)
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
	log.Printf("closed TCP socket %p\n", sock)
}

func (s Service) addTopic(t Topic) {
	s.topics[t.Id] = t.In
	// Notify Coordinator about created topic.
	s.Ipc.Write <- proto.Encode(proto.Message{
		Kind: proto.KindAdd,
		Id:   t.Id,
	})
	log.Printf("add topic %s\n", t.Id)
}

func (s Service) removeTopic(topicId string) {
	if _, ok := s.topics[topicId]; !ok {
		log.Printf("topic %s does not exist\n", topicId)
		return
	}
	delete(s.topics, topicId)
	// Notify Coordinator about removed topic.
	s.Ipc.Write <- proto.Encode(proto.Message{
		Kind: proto.KindRemove,
		Id:   topicId,
	})
	log.Printf("remove topic %s\n", topicId)
}

// WARN: It's the Coordinator's responsibility to validate messages that are comming
// from players.
func (s Service) route(encoded []byte) {
	// It is mandatory for every message to include an ID.
	topicId := proto.DecodeId(encoded[1:])
	// Immediately discard the message if the topic doesn't exist.
	in, ok := s.topics[topicId]
	if !ok {
		log.Printf("message to non-existing handler %s\n", topicId)
		return
	}

	m := proto.Message{
		Kind: encoded[0],
	}

	switch encoded[0] { // MessageKind.
	// For those messages, the payload contains the handler URL.
	// So the message must be simply forwarded to the named handler.
	case proto.KindJoin, proto.KindLeave:
		// Payload of the message contains the player ID.
		m.Payload = encoded[len(topicId)+1:]
	}

	// Forward the message to topic.
	in <- m
}

// cleanup is called only in case the server crushes. It is needed to gracefully
// terminate TCP connections without losing packets.
func (s Service) cleanup() {
	if err := s.listener.Close(); err != nil {
		log.Printf("cannot close TCP listener: %v\n", err)
	}
	// TODO: close all remaining connections.
}
