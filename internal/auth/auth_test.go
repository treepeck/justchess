// Tests from this package must be executed only when the testdb service is up and
// running.
package auth

import (
	"justchess/internal/db"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func initServiceOrPanic() Service {
	pool, err := db.OpenDB(os.Getenv("MYSQL_TEST_URL"))
	if err != nil {
		panic(err)
	}

	return NewService(db.NewRepo(pool))
}

func TestSignup(t *testing.T) {
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

		s.signup(rec, req)

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

func TestSignin(t *testing.T) {
	s := initServiceOrPanic()

	testcases := []struct {
		name         string
		formEmail    string
		formPassword string
		expectedCode int
	}{
		{"signin valid player", "magnus@carlsen.com", "carlsen", 200},
		{"signin valid player second session", "magnus@carlsen.com", "carlsen", 200},
		{"signin invalid email", "", "carlsen", 400},
		{"signin invalid password", "magnus@carlsen.com", "", 400},
		{"signin incorrect email", "m@carlsen.com", "carlsen", 401},
		{"signin incorrect password", "magnus@carlsen.com", "incorrect", 401},
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

		s.signin(rec, req)

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

func BenchmarkSignup(b *testing.B) {

}

func BenchmarkSignin(b *testing.B) {

}
