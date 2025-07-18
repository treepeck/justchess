package auth

import (
	"justchess/pkg/db"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	env, err := os.ReadFile("../../.env")
	if err != nil {
		log.Fatalf(".env file cannot be read %v", err)
	}

	for line := range strings.SplitSeq(string(env), "\n") {
		// Skip empty lines and comments.
		if len(line) < 3 || line[0] == '#' {
			continue
		}

		pair := strings.SplitN(line, "=", 2)
		// Skip malformed variable.
		if len(pair) != 2 {
			continue
		}

		os.Setenv(pair[0], pair[1])
	}

	db.Open()
	defer db.Close()

	os.Exit(m.Run())
}

func TestCreateSessionUnauthorized(t *testing.T) {
	req := httptest.NewRequest("POST", "/auth/", nil)

	rec := httptest.NewRecorder()
	createSession(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Must return statusok")
	}

	if res.Cookies()[0].Name != "Authorization" {
		t.Fatalf("Must return Authorization cookie")
	}

	sid := res.Cookies()[0].Value
	if err := db.DeleteSession(sid); err != nil {
		t.Fatalf("Cannot delete test session %v", err)
	}
}

func TestSignBySession(t *testing.T) {
	// Insert test session.
	var sid, uid string
	if err := db.InsertSession().Scan(&sid, &uid); err != nil {
		t.Fatalf("Cannot create test record %v", err)
	}
	// Delete test session.
	defer func() {
		if err := db.DeleteSession(sid); err != nil {
			t.Fatalf("Cannot delete test record %v", err)
		}
	}()

	req := httptest.NewRequest("GET", "/auth/", nil)
	req.AddCookie(&http.Cookie{
		Name:     "Authorization",
		Value:    sid,
		Path:     "/",
		MaxAge:   10,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	rec := httptest.NewRecorder()
	signBySession(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Must return statusok")
	}

	raw := make([]byte, 36)
	_, err := res.Body.Read(raw)
	if err != nil {
		t.Fatalf("Cannot read uid from body %v", err)
	}

	if string(raw) != uid {
		t.Fatalf("Expected %s got %s", uid, string(raw))
	}
}
