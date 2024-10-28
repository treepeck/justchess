package ws

import (
	"chess-api/db"
	"chess-api/models/game/enums"
	"chess-api/repository"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// createTestServer is a helper function which creates a test websocket server.
// Created server should be closed after executing tests by calling the server.Close().
// Database should alse be closed by calling the db.CloseDatabase func.
func createTestServer() *httptest.Server {
	// load env
	err := godotenv.Load("./../.env")
	if err != nil {
		return nil
	}

	err = db.OpenDatabase("./../db/schema.sql")
	if err != nil {
		return nil
	}

	m := NewManager()

	return httptest.NewServer(http.HandlerFunc(m.HandleConnection))
}

// Ignore logs from writeEvents and readEvents functions during tests.
func TestHandleConnection(t *testing.T) {
	s := createTestServer()
	defer db.CloseDatabase()
	if s == nil { // server wasnt created
		return
	}
	defer s.Close()

	testcases := []struct {
		name               string
		userId             string
		expectedStatusCode int
	}{
		{
			"add authorized user",
			"6635826b-5307-4155-9ef6-bff8bb8c3fc4",
			http.StatusSwitchingProtocols,
		},
		{
			"add unauthorized user",
			"fdbb963c-132b-4afb-9796-935f911f276b",
			http.StatusUnauthorized,
		},
		{
			"add user with malformed id",
			"asdsd",
			http.StatusBadRequest,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// url is generated dynamically
			url := "ws" + strings.TrimPrefix(s.URL, "http") +
				"/ws?id=" + tc.userId
			// connect to a Manager
			c, r, _ := websocket.DefaultDialer.Dial(url, nil)
			if c != nil {
				defer c.Close()
			}

			if r.StatusCode != tc.expectedStatusCode {
				t.Errorf("expected status: %d, got: %d", tc.expectedStatusCode, r.StatusCode)
			}
		})
	}
}

func TestAddRoom(t *testing.T) {
	s := createTestServer()
	defer db.CloseDatabase()
	if s == nil { // server wasnt created
		return
	}
	defer s.Close()

	// url is generated dynamically
	url := "ws" + strings.TrimPrefix(s.URL, "http") +
		"/ws?id=" + "6635826b-5307-4155-9ef6-bff8bb8c3fc4"
	// connect to a Manager
	c, _, _ := websocket.DefaultDialer.Dial(url, nil)
	if c == nil {
		return
	}
	defer c.Close()

	id, _ := uuid.Parse("6635826b-5307-4155-9ef6-bff8bb8c3fc4")
	user := repository.FindUserById(id)
	if user == nil {
		return
	}

	testcases := []struct {
		name        string
		cr          CreateRoomDTO
		expectedRes Event
	}{
		{
			"add room",
			CreateRoomDTO{
				Control: enums.Blitz,
				Bonus:   2,
				Owner:   *user,
			},
			Event{
				Action:  REDIRECT,
				Payload: nil,
			},
		},
		{
			"add multiple rooms",
			CreateRoomDTO{},
			Event{
				Action:  CREATE_ROOM_ERR,
				Payload: nil,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			payload, _ := json.Marshal(tc.cr)
			e := Event{
				Action:  CREATE_ROOM,
				Payload: payload,
			}
			c.WriteJSON(e)

			if tc.name == "add room" {
				c.ReadMessage() // skip first message
			}

			var got Event
			err := c.ReadJSON(&got)
			if err != nil {
				t.Fatalf("failed to read websocket response: %v", err)
			}
			if got.Action != tc.expectedRes.Action {
				t.Errorf("expected action: %s, got: %s", tc.expectedRes.Action, got.Action)
			}
		})
	}
}
