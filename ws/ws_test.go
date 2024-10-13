package ws

import (
	"chess-api/db"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Ignore logs from writeEvents and readEvents functions during tests.
func TestHandleConnection(t *testing.T) {
	// load env
	err := godotenv.Load("./../.env")
	if err != nil {
		return
	}

	err = db.OpenDatabase("./../db/schema.sql")
	if err != nil {
		return
	}
	defer db.CloseDatabase()

	m := NewManager()

	s := httptest.NewServer(http.HandlerFunc(m.HandleConnection))
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
