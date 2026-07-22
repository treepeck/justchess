package ws

import (
	"justchess/internal/talk"
	"log"
)

const msgBan string = "You have been banned for suspicious activity"

type room struct {
	channels   talk.GameChannels
	clients    map[string]*client
	register   chan registrant
	unregister chan string
}

func initRoom(channels talk.GameChannels) room {
	r := room{
		channels:   channels,
		clients:    make(map[string]*client, 2),
		register:   make(chan registrant),
		unregister: make(chan string),
	}
	go r.listen()
	return r
}

func (r room) listen() {
	for {
		select {
		case reg := <-r.register:
			r.onRegister(reg)
		case id := <-r.unregister:
			r.onUnregister(id)
		case msg := <-r.channels.Out:
			r.broadcast(msg)
		case id := <-r.channels.Ban:
			r.onBan(id)
		}
	}
}

func (r room) onRegister(reg registrant) {
	// Decline the request if the client is already registered.
	if _, exists := r.clients[reg.id]; exists || len(r.clients) == clientsThreshold {
		reg.err <- errAlreadyRegistered
		return
	}
	defer func() {
		reg.err <- nil
	}()

	conn, err := upgrader.Upgrade(reg.res, reg.req, nil)
	if err != nil {
		// Upgrader writes the response, so simply return here.
		return
	}
	c := newClient(conn, r.channels.In, r.unregister)
	go c.read(reg.id)
	go c.write()
	r.clients[reg.id] = c

	msg, err := talk.JSON(talk.MessageJoin, reg.id)
	if err != nil {
		log.Print(err)
		return
	}
	r.broadcast(msg)
}

func (r room) onUnregister(id string) {
	if _, exists := r.clients[id]; !exists {
		return
	}
	delete(r.clients, id)
	msg, err := talk.JSON(talk.MessageLeave, id)
	if err != nil {
		log.Print(err)
		return
	}
	r.broadcast(msg)
}

func (r room) onBan(id string) {
	c, exists := r.clients[id]
	if !exists {
		return
	}
	msg, _ := talk.JSON(talk.MessageError, msgBan)
	c.send <- msg
}

func (r room) broadcast(msg []byte) {
	for _, c := range r.clients {
		c.send <- msg
	}
}
