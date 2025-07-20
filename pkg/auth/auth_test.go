package auth

import (
	"bytes"
	"encoding/json"
	"justchess/pkg/db"
	"justchess/pkg/env"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	env.Load("../../.env")

	db.Open()
	defer db.Close()

	os.Exit(m.Run())
}

func TestSignup(t *testing.T) {
	testcases := []struct {
		name           string
		dto            signupDTO
		expectedStatus int
	}{
		{"invalid name", signupDTO{Name: "", Email: "test@test.com", Password: "test1"}, 406},
		{"invalid email", signupDTO{Name: "test", Email: "", Password: "test1"}, 406},
		{"short pwd", signupDTO{Name: "test", Email: "test@test.com", Password: ""}, 406},
		{"long pwd", signupDTO{Name: "test", Email: "test@test.com",
			Password: "1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"},
			406,
		},
	}

	for _, tc := range testcases {
		body, err := json.Marshal(tc.dto)
		if err != nil {
			t.Fatalf("Cannot Marshal dto %v", err)
		}

		req := httptest.NewRequest("POST", "/auth/signup", bytes.NewReader(body))
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
		dto            signinDTO
		expectedStatus int
	}{
		{"invalid email", signinDTO{Email: "", Password: "test1"}, 406},
		{"short pwd", signinDTO{Email: "test@test.com", Password: ""}, 406},
		{"long pwd", signinDTO{Email: "test@test.com",
			Password: "1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"},
			406,
		},
	}

	for _, tc := range testcases {
		body, err := json.Marshal(tc.dto)
		if err != nil {
			t.Fatalf("Cannot Marshal dto %v", err)
		}

		req := httptest.NewRequest("POST", "/auth/signin", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		signin(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		if res.StatusCode != tc.expectedStatus {
			t.Fatalf("Test %s failed: expeted %d got %d", tc.name, tc.expectedStatus, res.StatusCode)
		}
	}
}
