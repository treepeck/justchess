package auth

import (
	"encoding/json"
	"justchess/pkg/models/user"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func TestHandleCreateGuest(t *testing.T) {
	err := loadEnv()
	if err != nil {
		t.Fatalf("cannot read environment variables: %v", err)
	}

	req, err := http.NewRequest("GET", "localhost:3502/auth/guest", nil)
	if err != nil {
		t.Fatalf("cannot create request: %v", err)
	}

	rec := httptest.NewRecorder()
	handleCreateGuest(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status: %d, got: %d", http.StatusOK, res.StatusCode)
	}

	var g user.Guest
	err = json.NewDecoder(res.Body).Decode(&g)
	if err != nil {
		t.Fatalf("cannot decode response body: %v", err)
	}

	hasCookie := false
	for _, c := range res.Cookies() {
		if c.Name == "Authorization" {
			hasCookie = true
		}
	}
	if !hasCookie {
		t.Fatalf("Authorization cookie has not been set")
	}
}

func TestHandleRefreshTokens(t *testing.T) {
	err := loadEnv()
	if err != nil {
		t.Fatalf("cannot read environment variables: %v", err)
	}

	testcases := []struct {
		name           string
		refreshToken   string
		expectedStatus int
		expectCookie   bool
	}{
		{
			"valid_refresh",
			"",
			200,
			true,
		},
		{
			"invalid_token",
			"",
			401,
			false,
		},
	}

	for _, tc := range testcases {
		if tc.name == "valid_refresh" {
			rt, err := generateToken(uuid.New(), "RTS", time.Minute*5)
			if err != nil {
				t.Fatalf("cannot generate token: %v", err)
			}
			tc.refreshToken = rt
		}

		req, err := http.NewRequest("GET", "localhost:3502/auth/tokens", nil)
		if err != nil {
			t.Fatalf("cannot create request: %v", err)
		}

		c := http.Cookie{
			Name:     "Authorization",
			Value:    "Bearer " + tc.refreshToken,
			MaxAge:   100000,
			HttpOnly: true,
		}
		req.AddCookie(&c)

		rec := httptest.NewRecorder()
		handleRefreshTokens(rec, req)

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

func TestHandleGetUserByRefreshToken(t *testing.T) {
	err := loadEnv()
	if err != nil {
		t.Fatalf("cannot read environment variables: %v", err)
	}

	testcases := []struct {
		name           string
		refreshToken   string
		expectedStatus int
	}{
		{
			"valid_refresh_guest",
			"",
			200,
		},
		{
			"invalid_refrsh",
			"",
			401,
		},
	}

	for _, tc := range testcases {
		if tc.name[0:5] == "valid" {
			rt, err := generateToken(uuid.New(), "RTS", time.Minute*5)
			if err != nil {
				t.Fatalf("cannot generate token: %v", err)
			}
			tc.refreshToken = rt
		}

		req, err := http.NewRequest("GET", "localhost:3502/auth/me", nil)
		if err != nil {
			t.Fatalf("cannot create request: %v", err)
		}

		c := http.Cookie{
			Name:     "Authorization",
			Value:    "Bearer " + tc.refreshToken,
			MaxAge:   100000,
			HttpOnly: true,
		}
		req.AddCookie(&c)

		rec := httptest.NewRecorder()
		handleGetUserByRefreshToken(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		if res.StatusCode != tc.expectedStatus {
			t.Fatalf("expected status: %d, got: %d", tc.expectedStatus, res.StatusCode)
		}
	}
}

func loadEnv() error {
	return godotenv.Load("../../.env")
}
