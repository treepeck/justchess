package user

import (
	"bufio"
	"bytes"
	"encoding/json"
	"justchess/pkg/db"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCreateUserHandler(t *testing.T) {
	loadEnv()
	db.Open()
	defer db.Pool.Close()

	r := Register{
		Mail:     os.Getenv("SMTP_TEST_MAIL"),
		Name:     "test",
		Password: "1234567",
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Cannot Marshal request body: %v\n", err)
	}

	req, err := http.NewRequest("POST", "/user/", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("Cannot create request: %v\n", err)
	}

	rec := httptest.NewRecorder()
	CreateUserHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code: 200, got: %d\n", res.StatusCode)
	}

	// Delete the created unverified user.
	_, err = db.Pool.Exec("DELETE FROM unverified WHERE mail = $1;", r.Mail)
	if err != nil {
		t.Fatalf("cannot delete test user: %v\n", err)
	}
}

func TestVerifyHandler(t *testing.T) {
	loadEnv()
	db.Open()
	defer db.Pool.Close()

	r := Register{
		Mail:     "test@test",
		Name:     "1234567",
		Password: "1234567",
	}

	// First of all, insert test unverified record.
	id, err := insertUnverified(r)
	if err != nil {
		t.Fatalf("cannot insert unverified test record: %v\n", err)
	}

	req, err := http.NewRequest("GET", "/auth/verify?id="+id, nil)
	if err != nil {
		t.Fatalf("cannot create request: %v\n", err)
	}

	rec := httptest.NewRecorder()
	VerifyHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code: 200, got: %d\n", res.StatusCode)
	}

	// Delete the created user.
	_, err = db.Pool.Exec("DELETE FROM users WHERE mail = $1;", r.Mail)
	if err != nil {
		t.Fatalf("cannot delete test user: %v\n", err)
	}
}

// loadEnv reads test.env file.
// Accepted format for variable: KEY=VALUE
// Comments which begin with '#' and empty lines are skipped.
func loadEnv() {
	f, err := os.Open("../../test.env")
	if err != nil {
		log.Fatalf("cannot read .env file: %v\n", err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())

		// Skip empty lines and comments.
		if len(line) < 2 || line[0] == '#' {
			continue
		}

		pair := strings.SplitN(line, "=", 2)
		// Skip malformed variable.
		if len(pair) != 2 {
			continue
		}

		os.Setenv(pair[0], pair[1])
	}

	if err := s.Err(); err != nil {
		log.Fatalf("error while reading .env: %v\n", err)
	}
}
