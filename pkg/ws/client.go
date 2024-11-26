package ws

import (
	"encoding/json"
	"log/slog"
	"time"

	"justchess/pkg/models/game/enums"
	"justchess/pkg/models/game/helpers"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Deadline for the next pong message from peer.
	pongWait = 10 * time.Second
	// Client sends ping messages with the defined interval.
	// It must be less than pongWait.
	pingInterval = (pongWait * 9) / 10
)

// Client stores the connection and writes events by using a channel.
// The use of a channel is necessary, since whe connection supports
// only one concurrent write at a time.
type Client struct {
	Id               uuid.UUID
	conn             *websocket.Conn
	manager          *Manager
	writeEventBuffer chan Event
	currentRoom      *Room
}

// newClient creates a new client.
func newClient(conn *websocket.Conn, m *Manager, id uuid.UUID) *Client {
	return &Client{
		Id:               id,
		conn:             conn,
		manager:          m,
		writeEventBuffer: make(chan Event),
		currentRoom:      nil,
	}
}

// readEvents reads and handles all incoming messages (events) from the connection.
func (c *Client) readEvents() {
	defer func() {
		c.manager.unregister <- c
	}()

	c.conn.SetReadLimit(10000)
	// set the read deadline to limit inactive connections.
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		slog.Warn("error while setting the read deadline", "err", err)
		return
	}
	c.conn.SetPongHandler(c.pongHandler)

	// forever loop to read incomming Events aka (messages) from the peer.
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway,
				websocket.CloseNormalClosure) {
				slog.Warn("error while reading a message", "err", err)
			}
			break
		}

		var e Event
		err = json.Unmarshal(data, &e)
		if err != nil {
			slog.Warn("cannot Unmarshal event", "err", err)
			break
		}
		c.handleEvent(e)
	}
}

// writeEvents grabs the events from the writeEventBuffer channel
// and sends those events to the client.
func (c *Client) writeEvents() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.manager.unregister <- c
	}()
	// forever loop grabs the incomming events from a channel and writes them
	// through a connection.
	for {
		select {
		case e, ok := <-c.writeEventBuffer:
			c.conn.SetWriteDeadline(time.Now().Add(pingInterval))
			if !ok {
				slog.Info("connection closed")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, e.Marshal()); err != nil {
				slog.Warn("failed to write event", "err", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(pingInterval))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte("Ping")); err != nil {
				return
			}
		}
	}
}

// handleEvent handles the incomming events by calling corresponding functions.
func (c *Client) handleEvent(e Event) {
	switch e.Action {
	case CREATE_ROOM:
		c.handleCreateRoom(e.Payload)

	case JOIN_ROOM:
		c.handleJoinRoom(e.Payload)

	case LEAVE_ROOM:
		c.handleLeaveRoom()

	case GET_ROOMS:
		c.handleGetRooms()

	case GET_GAME:
		c.handleGetGame(e.Payload)

	case MOVE:
		c.handleMove(e.Payload)

	case SEND_MESSAGE:
		c.handleSendMessage(e.Payload)

	default:
		slog.Warn("event have unknown action", "action", e.Action)
	}
}

func (c *Client) handleCreateRoom(payload json.RawMessage) {
	var cr CreateRoomDTO
	err := json.Unmarshal(payload, &cr)
	if err != nil {
		slog.Warn("cannot Unmarshal CreateRoomDTO", "err", err)
		c.writeEventBuffer <- Event{
			Action:  CREATE_ROOM_ERR,
			Payload: nil,
		}
		return
	}

	if c.currentRoom != nil {
		slog.Info("cannot create multiple rooms")
		c.writeEventBuffer <- Event{
			Action:  CREATE_ROOM_ERR,
			Payload: nil,
		}
		return
	}
	r := newRoom(cr)
	c.currentRoom = r
	c.manager.add <- r
	c.sendEvent(REDIRECT, r.id)
}

func (c *Client) handleJoinRoom(payload json.RawMessage) {
	roomId, err := uuid.Parse(string(payload))
	if err != nil {
		slog.Warn("cannot parse roomId", "err", err)
		c.sendError(UNPROCESSABLE_ENTITY)
		return
	}
	if r := c.manager.findRoomById(roomId); r != nil &&
		c.currentRoom == nil {
		c.sendEvent(REDIRECT, r.id)
	}
}

func (c *Client) handleLeaveRoom() {
	if c.currentRoom != nil {
		c.currentRoom.unregister <- c
	}
}

// getGame sends the latest data about the specified game.
func (c *Client) handleGetGame(payload json.RawMessage) {
	roomId, err := uuid.Parse(string(payload))
	if err != nil {
		slog.Warn("cannot Parse roomId", "err", err)
		c.sendError(UNPROCESSABLE_ENTITY)
		return
	}
	if r := c.manager.findRoomById(roomId); r != nil {
		r.register <- c
		// } else if g := repository.FindGameById(roomId); g != nil {
		// 	c.sendEvent(MOVES, g.Moves)
		// 	c.sendEvent(GAME_INFO, g)
		// 	c.sendEvent(GAME_INFO, g.Result)
		// }
	}
}

func (c *Client) handleMove(payload json.RawMessage) {
	var m helpers.Move
	err := json.Unmarshal(payload, &m)
	if err != nil {
		slog.Warn("cannot Unmarshal MoveDTO", "err", err)
		return
	}

	if c.currentRoom != nil {
		c.currentRoom.handleTakeMove(m, c)
	}
}

// handleGetRooms sends the current availible rooms one by one.
// There can be a lot of rooms, so they can`t be send as a single message.
func (c *Client) handleGetRooms() {
	for r := range c.manager.rooms {
		if r.game.Status == enums.Waiting {
			payload, err := json.Marshal(r)
			if err != nil {
				slog.Warn("cannot Marshal Room", "err", err)
				continue
			}
			e := Event{
				Action:  ADD_ROOM,
				Payload: payload,
			}
			c.writeEventBuffer <- e
		}
	}
}

func (c *Client) handleSendMessage(payload json.RawMessage) {
	if c.currentRoom != nil {
		c.currentRoom.broadcastChatMessage(payload, c.Id)
	}
}

func (c *Client) sendEvent(a string, pData any) {
	p, err := json.Marshal(pData)
	if err != nil {
		slog.Warn("cannot send event: "+a, "err", err)
		return
	}
	e := Event{
		Action:  a,
		Payload: p,
	}
	c.writeEventBuffer <- e
}

// sendLastMove serializes move into Long Algebraic Notation and sends it
// with the timer left duration.
func (c *Client) sendLastMove(m helpers.Move, pt enums.PieceType) {
	type lastMoveDTO struct {
		UCI        string            `json:"uci"`
		LAN        string            `json:"lan"` // TODO: replace with SAN
		FEN        string            `json:"fen"`
		ValidMoves map[string]string `json:"vm"`
		TimeLeft   time.Duration     `json:"timeLeft"`
	}
	// convert map[helpers.Pos][]helpers.PossibleMove to JS Object with
	// string keys and string values.
	vm := make(map[string]string)
	for pos, moves := range c.currentRoom.game.CurrentValidMoves {
		vm[pos.String()] = ""
		for _, m := range moves {
			vm[pos.String()] += m.To.String()
		}
	}

	lm := lastMoveDTO{
		UCI:        m.From.String() + m.To.String() + m.PromotionPayload.String(),
		LAN:        m.ToLAN(pt),
		FEN:        c.currentRoom.game.ToFEN(),
		ValidMoves: vm,
		TimeLeft:   m.TimeLeft,
	}

	p, err := json.Marshal(lm)
	if err != nil {
		slog.Warn("cannot Marshal last move", "err", err)
		return
	}
	e := Event{
		Action:  LAST_MOVE,
		Payload: p,
	}
	c.writeEventBuffer <- e
}

// sendError sends an emerged error as Event type.
func (c *Client) sendError(errName string) {
	e := Event{
		Action:  errName,
		Payload: nil,
	}
	c.writeEventBuffer <- e
}

// pongHandler adds pongWait to the read deadline for the next pong message.
func (c *Client) pongHandler(_ string) error {
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}
