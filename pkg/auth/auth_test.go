package auth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"justchess/pkg/db"
	"justchess/pkg/user"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// All actions are applied to the test database 'justchess_test'.
// That db must have the same schema as a 'justchess'.
// All data will be automatically cleaned-up at the end of the tests.

// TestSignUpHandler will try send an email to the (SMTP_TEST_MAIL) address
// parsed from the test.env file.
func TestSignUpHandler(t *testing.T) {
	loadEnv()
	db.Open()
	defer func() {
		// Delete the created unverified user.
		db.Pool.Exec("DELETE FROM unverified WHERE mail = $1;", os.Getenv("SMTP_TEST_MAIL"))
		db.Pool.Close()
	}()

	testcases := []struct {
		name               string
		register           user.Register
		expectedStatusCode int
	}{
		{
			"valid_user",
			user.Register{Mail: os.Getenv("SMTP_TEST_MAIL"), Name: "test1", Password: "doesnt_matter"},
			200,
		},
		{
			"invalid_email",
			user.Register{Mail: "doesnt_matter", Name: "test2", Password: "doesnt_matter"},
			400,
		},
		{"bad_request", user.Register{}, 400},
	}

	for _, tc := range testcases {
		data, err := json.Marshal(tc.register)
		if err != nil {
			t.Fatalf("Cannot Marshal request body: %v\n", err)
		}

		req, err := http.NewRequest("POST", "/auth/", bytes.NewBuffer(data))
		if err != nil {
			t.Fatalf("Cannot create request: %v\n", err)
		}

		rec := httptest.NewRecorder()
		SignUpHandler(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		if res.StatusCode != tc.expectedStatusCode {
			t.Fatalf("expected status code: %d, got: %d\n", tc.expectedStatusCode, res.StatusCode)
		}
	}
}

func TestVerifyMailHandler(t *testing.T) {
	loadEnv()
	db.Open()
	defer func() {
		// Delete the created verified user.
		db.Pool.Exec("DELETE FROM users WHERE user_name = $1;", "test")
		db.Pool.Close()
	}()

	tx, err := db.Pool.Begin()
	if err != nil {
		t.Fatalf("cannot begin transaction: %v\n", err)
	}
	id, err := user.InsertUnverified(user.Register{Mail: "test", Name: "test", Password: "test123"}, tx)
	if err != nil {
		t.Fatalf("cannot create unverified user: %v\n", err)
	}
	tx.Commit()

	testcases := []struct {
		name               string
		id                 string
		expectedStatusCode int
	}{
		{"valid_verify", id, 200},
		{"bad_request", "", 400},
		{"non_existed_user", uuid.New().String(), 401},
	}

	for _, tc := range testcases {
		req, err := http.NewRequest("POST", "/auth/veify?id="+tc.id, nil)
		if err != nil {
			t.Fatalf("Cannot create request: %v\n", err)
		}

		rec := httptest.NewRecorder()
		VerifyMailHandler(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		if res.StatusCode != tc.expectedStatusCode {
			t.Fatalf("expected status code: %d, got: %d\n", tc.expectedStatusCode, res.StatusCode)
		}
	}
}

func TestRefreshHandler(t *testing.T) {
	loadEnv()
	db.Open()
	defer db.Pool.Close()

	id := uuid.New()
	defer func() {
		// Delete the created verified user.
		db.Pool.Exec("DELETE FROM users WHERE id = $1;", id.String())
		db.Pool.Close()
	}()

	tx, err := db.Pool.Begin()
	if err != nil {
		t.Fatalf("cannot begin transaction: %v\n", err)
	}
	_, err = user.InsertUser(id.String(), user.Register{
		Mail: "test", Name: "test", Password: "test123",
	}, tx)
	if err != nil {
		t.Fatalf("cannot create verified user: %v\n", err)
	}
	tx.Commit()

	testcases := []struct {
		name           string
		expectedStatus int
		expectCookie   bool
	}{
		{
			"valid_refresh",
			200,
			true,
		},
		{
			"invalid_token",
			401,
			false,
		},
	}

	for _, tc := range testcases {
		refreshToken := ""
		if tc.name == "valid_refresh" {
			rt, err := generateToken(
				id,
				"test",
				os.Getenv("REFRESH_TOKEN_SECRET"),
				time.Minute*5,
			)
			if err != nil {
				t.Fatalf("cannot generate token: %v", err)
			}
			refreshToken = rt
		}

		req, err := http.NewRequest("GET", "localhost:3502/auth", nil)
		if err != nil {
			t.Fatalf("cannot create request: %v", err)
		}

		c := http.Cookie{
			Name:     "Authorization",
			Value:    "Bearer " + refreshToken,
			MaxAge:   100000,
			HttpOnly: true,
		}
		req.AddCookie(&c)

		rec := httptest.NewRecorder()
		RefreshHandler(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		if res.StatusCode != tc.expectedStatus {
			t.Fatalf("expected status: %d, got: %d", tc.expectedStatus, res.StatusCode)
		}

		hasCookie := false
		for _, c := range res.Cookies() {
			if c.Name == "Authorization" {
				hasCookie = true
			}
		}
		if hasCookie != tc.expectCookie {
			t.Fatalf("expect cookie: %v, got: %v", tc.expectCookie, hasCookie)
		}
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
