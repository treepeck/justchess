package auth

import (
	"justchess/internal/db"
	"justchess/internal/env"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	env.Load("../../.env")

	db.Open()
	defer db.Close()

	os.Exit(m.Run())
}

type form struct {
	name     string
	email    string
	password string
}

func TestSignup(t *testing.T) {
	testcases := []struct {
		name           string
		dto            form
		expectedStatus int
	}{
		{"invalid name", form{name: "", email: "test@test.com", password: "test1"}, 406},
		{"invalid email", form{name: "test", email: "", password: "test1"}, 406},
		{"short pwd", form{name: "test", email: "test@test.com", password: ""}, 406},
		{"long pwd", form{name: "test", email: "test@test.com",
			password: "1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"},
			406,
		},
	}

	for _, tc := range testcases {
		body := url.Values{}
		body.Add("name", tc.dto.name)
		body.Add("email", tc.dto.email)
		body.Add("password", tc.dto.password)

		req := httptest.NewRequest("POST", "/auth/signup", strings.NewReader(body.Encode()))
		rec := httptest.NewRecorder()
		signup(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		if res.StatusCode != tc.expectedStatus {
			t.Fatalf("Test %s failed: expeted %d got %d", tc.name, tc.expectedStatus, res.StatusCode)
		}
	}
}

func TestSignin(t *testing.T) {
	testcases := []struct {
		name           string
		dto            form
		expectedStatus int
	}{
		{"invalid email", form{email: "", password: "test1"}, 406},
		{"short pwd", form{email: "test@test.com", password: ""}, 406},
		{"long pwd", form{email: "test@test.com",
			password: "1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"},
			406,
		},
	}

	for _, tc := range testcases {
		body := url.Values{}
		body.Add("email", tc.dto.email)
		body.Add("password", tc.dto.password)

		req := httptest.NewRequest("POST", "/auth/signin", strings.NewReader(body.Encode()))
		rec := httptest.NewRecorder()
		signin(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		if res.StatusCode != tc.expectedStatus {
			t.Fatalf("Test %s failed: expeted %d got %d", tc.name, tc.expectedStatus, res.StatusCode)
		}
	}
}
