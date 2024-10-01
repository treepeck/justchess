package ws

import (
	"chess-api/models"
	"chess-api/models/helpers"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type Client struct {
	User             models.User `json:"user"`
	conn             *websocket.Conn
	manager          *Manager
	writeEventBuffer chan Event
	currentRoomId    uuid.UUID
}

func newClient(conn *websocket.Conn, m *Manager, user models.User) *Client {
	return &Client{
		User:             user,
		conn:             conn,
		manager:          m,
		writeEventBuffer: make(chan Event),
		currentRoomId:    uuid.Nil,
	}
}

// Reads all incoming messages from the connection.
func (c *Client) readEvents() {
	fn := slog.String("func", "hub.readEvents")
	defer c.manager.removeClient(c)

	// set the read deadline to limit inactive connections
	c.conn.SetReadLimit(10000)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		slog.Warn("error while setting the read deadline", fn, "err", err)
		return
	}
	c.conn.SetPongHandler(c.pongHandler)

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var e Event
		err = json.Unmarshal(data, &e)
		if err != nil {
			slog.Warn("cannot Unmarshal event", fn, "err", err)
			return
		}
		c.handleEvent(e)
	}
}

func (c *Client) writeEvents() {
	fn := slog.String("func", "hub.writeEvents")
	defer c.manager.removeClient(c)

	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case e, ok := <-c.writeEventBuffer:
			c.conn.SetWriteDeadline(time.Now().Add(pingInterval))
			if !ok {
				slog.Info("connection closed", fn)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, e.Marshal()); err != nil {
				slog.Warn("failed to write event", fn, "err", err)
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

func (c *Client) pongHandler(_ string) error {
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}

func (c *Client) handleEvent(e Event) {
	fn := slog.String("func", "handleEvent")

	switch e.Action {

	case GET_ROOMS:
		var gr GetRoomDTO
		err := json.Unmarshal(e.Payload, &gr)
		if err != nil {
			slog.Warn("cannot Unmarshal GetRoomDTO", fn, "err", err)
			return
		}

		e := Event{
			Action: UPDATE_ROOMS,
		}
		var rooms []byte
		if gr.Bonus == "All" && gr.Control == "All" {
			rooms, err = json.Marshal(c.manager.roomController.FindAvailible())
		} else {
			rooms, err = json.Marshal(c.manager.roomController.FilterRooms(gr))
		}
		if err != nil {
			slog.Warn("cannot Marshal rooms", fn, "err", err)
			return
		}
		e.Payload = rooms
		c.writeEventBuffer <- e

	case CREATE_ROOM:
		var cr CreateRoomDTO
		err := json.Unmarshal(e.Payload, &cr)
		if err != nil {
			slog.Warn("cannot Unmarshal CreateRoomDTO", fn, "err", err)
			return
		}

		if r := c.manager.roomController.FindByOwnerId(cr.Owner.Id); r != nil {
			slog.Info("cannot create multiple rooms", fn)
			return
		}
		c.manager.createRoom(cr, c)

	case JOIN_ROOM:
		var idStr string
		json.Unmarshal(e.Payload, &idStr)
		roomId, err := uuid.Parse(idStr)
		if err != nil {
			slog.Warn("cannot parse roomId", fn, "err", err)
			return
		}
		if r := c.manager.roomController.FindById(roomId); r != nil &&
			c.currentRoomId == uuid.Nil {
			r.AddPlayer(c)
			c.changeRoom(r.Id)
		}

	case GET_GAME:
		var idStr string
		json.Unmarshal(e.Payload, &idStr)
		roomId, err := uuid.Parse(idStr)
		if err == nil && c.currentRoomId == roomId {
			if r := c.manager.roomController.FindById(roomId); r != nil {
				r.HandleGetGame(c)
			}
		}

	case MOVE:
		if c.currentRoomId != uuid.Nil {
			if r := c.manager.roomController.FindById(c.currentRoomId); r != nil {
				var move helpers.MoveDTO
				err := json.Unmarshal(e.Payload, &move)
				if err != nil {
					slog.Warn("cannot Unmsrshal position", fn, "err", err)
					return
				}
				r.HandleTakeMove(move, c)
			}
		}

	default:
		slog.Warn("event have unknown action", fn, "action", e.Action)
	}
}

func (c *Client) changeRoom(roomId uuid.UUID) {
	if c.currentRoomId == uuid.Nil {
		p, _ := json.Marshal(roomId.String())
		e := Event{
			Action:  CHANGE_ROOM,
			Payload: p,
		}
		c.currentRoomId = roomId
		c.writeEventBuffer <- e
	}
}
