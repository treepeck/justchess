/*
Tests from this package must be executed only when the testdb service is up and
running.
*/
package auth_test

import (
	"justchess/internal/auth"
	"justchess/internal/db"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func initServiceOrPanic() *auth.Service {
	pool, err := db.OpenDB(os.Getenv("MYSQL_TEST_URL"))
	if err != nil {
		panic(err)
	}

	return auth.NewService(db.NewRepo(pool))
}

func TestHandleSignup(t *testing.T) {
	s := initServiceOrPanic()

	testcases := []struct {
		name         string
		formName     string
		formEmail    string
		formPassword string
		expectedCode int
	}{
		{"signup valid player", "test", "test@test.com", "testtest", 200},
		{"signup duplicate name", "test", "test2@test.com", "testtest", 409},
		{"signup duplicate email", "test2", "test@test.com", "testtest", 409},
		{"signup invalid name", "1", "test3@test.com", "testtest", 406},
		{"signup invalid email", "test3", "2@.com", "testtest", 406},
		{"signup invalid password", "test3", "2@.com", "sd", 406},
	}

	for _, tc := range testcases {
		body := url.Values{}
		body.Set("name", tc.formName)
		body.Set("email", tc.formEmail)
		body.Set("password", tc.formPassword)

		req := httptest.NewRequest(
			"POST",
			"/auth/signup",
			strings.NewReader(body.Encode()),
		)

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		s.HandleSignup(rec, req)

		res := rec.Result()
		res.Body.Close()

		if res.StatusCode != tc.expectedCode {
			t.Fatalf(
				"%s failed: expected %d got %d",
				tc.name, tc.expectedCode, res.StatusCode,
			)
		}
	}
}

func TestHandleSignin(t *testing.T) {
	s := initServiceOrPanic()

	testcases := []struct {
		name         string
		formEmail    string
		formPassword string
		expectedCode int
	}{
		{"signin valid player", "magnus@carlsen.com", "carlsen", 200},
		{"signin duplicate", "magnus@carlsen.com", "carlsen", 409},
		{"signin invalid email", "", "carlsen", 400},
		{"signin invalid password", "magnus@carlsen.com", "", 400},
		{"signin incorrect email", "m@carlsen.com", "carlsen", 406},
		{"signin incorrect password", "magnus@carlsen.com", "incorrect", 406},
	}

	for _, tc := range testcases {
		body := url.Values{}
		body.Set("email", tc.formEmail)
		body.Set("password", tc.formPassword)

		req := httptest.NewRequest(
			"POST",
			"/auth/signin",
			strings.NewReader(body.Encode()),
		)

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		s.HandleSignin(rec, req)

		res := rec.Result()
		res.Body.Close()

		if res.StatusCode != tc.expectedCode {
			t.Fatalf(
				"%s failed: expected %d got %d",
				tc.name, tc.expectedCode, res.StatusCode,
			)
		}
	}
}

func TestHandleVerify(t *testing.T) {
	s := initServiceOrPanic()

	testcases := []struct {
		name         string
		sessionId    string
		expectedCode int
	}{
		{"verify valid player", "57w_sbICMc9znzXepVw2RskBDg_W94H1", 200},
		{"verify missing session", "", 400},
		{"verify missing player", "MIS_sbICMc9znzXepVw2RskBDg_W94H1", 401},
	}

	for _, tc := range testcases {
		req := httptest.NewRequest(
			"POST",
			"/auth/verify",
			strings.NewReader(tc.sessionId),
		)

		rec := httptest.NewRecorder()

		s.HandleVerify(rec, req)

		res := rec.Result()
		res.Body.Close()

		if res.StatusCode != tc.expectedCode {
			t.Fatalf(
				"%s failed: expected %d got %d",
				tc.name, tc.expectedCode, res.StatusCode,
			)
		}
	}
}
