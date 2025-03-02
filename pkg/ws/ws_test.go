package ws

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func TestHandleNewConnection(t *testing.T) {
	server, r := newTestServer()
	defer func() {
		server.Close()
		close(r.register)
	}()

	// Convert HTTP test server URL to WebSocket URL.
	wsURL := "ws" + server.URL[4:] + "/ws"

	// Test bad connection with the invalid JWT.
	_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatalf("should not be able to connect\n")
	}

	// Test valid connection.
	accessToken, err := generateToken(uuid.New(), os.Args[1], time.Minute*5)
	if err != nil {
		t.Fatalf("cannot generate token: %v\n", err)
	}
	wsURL += "?access=" + accessToken

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("cannot create connection: %v\n", err)
	}
	conn.Close()
}

func TestPsuedoGame(t *testing.T) {
	server, r := newTestServer()
	defer func() {
		server.Close()
		close(r.register)
	}()

	// Connect first player.
	accessToken, err := generateToken(uuid.New(), os.Args[1], time.Minute*5)
	if err != nil {
		t.Fatalf("cannot generate token: %v\n", err)
	}

	wsURL := "ws" + server.URL[4:] + "/ws?access=" + accessToken
	first, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("cannot create connection: %v\n", err)
	}
	defer first.Close()

	// Connect second player.
	accessToken, err = generateToken(uuid.New(), os.Args[1], time.Minute*5)
	if err != nil {
		t.Fatalf("cannot generate token: %v\n", err)
	}

	wsURL = "ws" + server.URL[4:] + "/ws?access=" + accessToken
	second, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("cannot create connection: %v\n", err)
	}
	defer second.Close()

	// Give some time for the room to respond to the request.
	time.Sleep(1 * time.Second)

	// At this moment, both clients are connected, so the room status must be IN_PROGRESS.
	if r.status != IN_PROGRESS {
		t.Fatalf("game has not begun.\n")
	}
}

// Starts a test server which handles WebSocket connections.
func newTestServer() (*httptest.Server, *Room) {
	r := NewRoom(uuid.New(), 100, 100)
	go r.EventPump()

	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		r.HandleNewConnection(rw, req)
	})), r
}

// Generates JWT.
func generateToken(id uuid.UUID, secret string,
	d time.Duration) (string, error) {
	unsigned := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   id.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(d)),
	})
	return unsigned.SignedString([]byte(secret))
}
