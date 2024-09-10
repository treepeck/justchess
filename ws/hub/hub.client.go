package hub

import (
	"chess-api/models"
	"chess-api/ws/event"
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

type hubClient struct {
	user             models.User
	conn             *websocket.Conn
	manager          *HubManager
	writeEventBuffer chan event.HubEvent
}

func newClient(conn *websocket.Conn, m *HubManager, user models.User) *hubClient {
	return &hubClient{
		user:             user,
		conn:             conn,
		manager:          m,
		writeEventBuffer: make(chan event.HubEvent),
	}
}

// Reads all incoming messages from the connection.
func (c *hubClient) readEvents() {
	fn := slog.String("func", "readEvents")

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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				slog.Warn("abnormal websocket closure", fn, "err", err)
			}
			return
		}

		var e event.HubEvent
		err = json.Unmarshal(data, &e)
		if err != nil {
			slog.Warn("cannot Unmarshal event", fn, "err", err)
			return
		}
		c.handleEvent(e)
	}
}

func (c *hubClient) writeEvents() {
	fn := slog.String("func", "writeEvents")

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

func (c *hubClient) pongHandler(_ string) error {
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}

func (c *hubClient) handleEvent(e event.HubEvent) {
	fn := slog.String("func", "handleEvent")

	switch e.Action {

	case event.GET_ROOMS:
		rooms, err := json.Marshal(c.manager.getAllRooms())
		if err != nil {
			slog.Warn("cannot Marshal rooms", fn, "err", err)
			return
		}
		e.Action = event.UPDATE_ROOMS
		e.Payload = rooms
		c.writeEventBuffer <- e

	case event.CREATE_ROOM:
		// TODO: add a check to block multiple rooms creating by a single user
		var cr createRoomDTO
		err := json.Unmarshal(e.Payload, &cr)
		if err != nil {
			slog.Warn("canot Unmarshal CreateRoomDTO", fn, "err", err)
			return
		}

		r := c.manager.createRoom(cr)
		p, _ := json.Marshal(r.Id.String())

		e := event.HubEvent{
			Action:  event.CHANGE_ROOM,
			Payload: p,
		}
		c.writeEventBuffer <- e

	case event.JOIN_ROOM:
		var idStr string
		json.Unmarshal(e.Payload, &idStr)
		roomId, err := uuid.Parse(idStr)
		if err != nil {
			slog.Warn("cannot parse roomId ", fn, "err", err)
			return
		}

		if r := c.manager.findRoomById(roomId); r != nil {
			e := event.HubEvent{
				Action:  event.CHANGE_ROOM,
				Payload: e.Payload,
			}
			c.writeEventBuffer <- e
		}

	default:
		slog.Warn("event have unknown action", fn, "action", e.Action)
	}
}
