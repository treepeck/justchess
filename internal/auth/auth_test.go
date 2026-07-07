package auth

import (
	"errors"
	"justchess/internal/db"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type mockAuthRepo struct {
}

func (r mockAuthRepo) InsertPlayer(id string, d db.SignupData) error {
	return nil
}

func (r mockAuthRepo) IsEmailUnique(email string) (bool, error) {
	if email == "not@unique.com" {
		return false, nil
	}
	return true, nil
}

func (r mockAuthRepo) SelectCredentialsByEmail(email string) (db.Credentials, error) {
	if email == "invalid@invalid.com" {
		return db.Credentials{}, errors.New("missing")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("valid"), bcrypt.DefaultCost)
	if email == "many@sessions.com" {
		return db.Credentials{
			Id:           "manySessions",
			PasswordHash: hash,
		}, nil
	}
	return db.Credentials{
		PasswordHash: hash,
	}, nil
}

func (r mockAuthRepo) SelectIdentityByEmail(email string) (db.Identity, error) {
	if email == "invalid@invalid.com" {
		return db.Identity{}, errors.New("unauthorized")
	}
	return db.Identity{}, nil
}

func (r mockAuthRepo) UpdatePasswordHash(id string, pwdHash []byte) error {
	return nil
}

func (r mockAuthRepo) InsertSignupToken(id string, d db.SignupData) error {
	return nil
}

func (r mockAuthRepo) SelectSignupDataByToken(id string) (db.SignupData, error) {
	if id == "valid" {
		return db.SignupData{}, nil
	}
	return db.SignupData{}, errors.New(id)
}

func (r mockAuthRepo) DeleteSignupToken(id string) error {
	return nil
}

func (r mockAuthRepo) InsertPasswordResetToken(id, playerId string, pwdHash []byte) error {
	if err := bcrypt.CompareHashAndPassword(pwdHash, []byte("duplicate")); err == nil {
		return errors.New("duplicate password")
	}
	return nil
}

func (r mockAuthRepo) SelectCredentialsByResetToken(id string) (db.Credentials, error) {
	if id == "valid" {
		return db.Credentials{}, nil
	}
	return db.Credentials{}, errors.New(id)
}

func (r mockAuthRepo) DeletePasswordResetToken(id string) error {
	return nil
}

func initServiceOrPanic() Service {
	cookieKey, err := ParseCookieKey(os.Getenv("COOKIE_KEY"))
	if err != nil {
		panic(err)
	}

	s := NewService(cookieKey, mockAuthRepo{})
	if err = s.ParseEmails("./templates/"); err != nil {
		panic(err)
	}
	return s
}

func TestSignup(t *testing.T) {
	s := initServiceOrPanic()

	cases := []struct {
		formName     string
		formEmail    string
		formPassword string
		expectedCode int
	}{
		{"valid", "valid@valid.com", "valid", http.StatusOK},
		{"missingEmail", "", "valid", http.StatusNotAcceptable},
		{"x", "small@name.com", "valid", http.StatusNotAcceptable},
		{"TOOOOOLONGNAMESFIDFNDSIFNODSNFSODNFDONFSDIONasdASDASDASDdDdDD", "valid@valid.com", "valid", http.StatusNotAcceptable},
		{"missingPassword", "valid@valid.com", "", http.StatusNotAcceptable},
		{"valid", "not@unique.com", "valid", http.StatusConflict},
	}

	for i, tc := range cases {
		body := url.Values{}
		body.Set("username", tc.formName)
		body.Set("email", tc.formEmail)
		body.Set("password", tc.formPassword)

		req := httptest.NewRequest("POST", "/", strings.NewReader(body.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		s.signup(rec, req)

		res := rec.Result()
		res.Body.Close()
		if res.StatusCode != tc.expectedCode {
			t.Fatalf("case %d failed: expected %d got %d", i, tc.expectedCode, res.StatusCode)
		}
	}
}

func TestSignin(t *testing.T) {
	s := initServiceOrPanic()

	cases := []struct {
		formEmail    string
		formPassword string
		expectedCode int
	}{
		{"valid@valid.com", "valid", http.StatusFound},
		{"", "valid", http.StatusBadRequest},
		{"valid@valid.com", "", http.StatusBadRequest},
		{"invalid@invalid.com", "valid", http.StatusUnauthorized},
		{"valid@valid.com", "invalid", http.StatusUnauthorized},
		{"many@sessions.com", "valid", http.StatusFound},
	}

	for i, tc := range cases {
		body := url.Values{}
		body.Set("email", tc.formEmail)
		body.Set("password", tc.formPassword)

		req := httptest.NewRequest("POST", "/", strings.NewReader(body.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		s.signin(rec, req)

		res := rec.Result()
		res.Body.Close()
		if res.StatusCode != tc.expectedCode {
			t.Fatalf("case %d failed: expected %d got %d", i, tc.expectedCode, res.StatusCode)
		}
	}
}

func TestResetPassword(t *testing.T) {
	s := initServiceOrPanic()

	cases := []struct {
		formEmail    string
		formPassword string
		expectedCode int
	}{
		{"valid@valid.com", "valid", http.StatusOK},
		{"valid@valid.com", "", http.StatusBadRequest},
		{"invalid@invalid.com", "valid", http.StatusUnauthorized},
		{"valid@valid.com", "duplicate", http.StatusConflict},
	}

	for i, tc := range cases {
		body := url.Values{}
		body.Set("email", tc.formEmail)
		body.Set("password", tc.formPassword)

		req := httptest.NewRequest("POST", "/", strings.NewReader(body.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		s.resetPassword(rec, req)

		res := rec.Result()
		res.Body.Close()
		if res.StatusCode != tc.expectedCode {
			t.Fatalf("case %d failed: expected %d got %d", i, tc.expectedCode, res.StatusCode)
		}
	}
}

func TestConfirmSignup(t *testing.T) {
	s := initServiceOrPanic()

	cases := []struct {
		token          string
		expectedStatus int
	}{
		{"valid", http.StatusFound},
		{"", http.StatusNotFound},
		{"invalid", http.StatusNotFound},
	}

	for i, tc := range cases {
		req := httptest.NewRequest("POST", "/", nil)
		req.SetPathValue("token", tc.token)

		rec := httptest.NewRecorder()
		s.confirmSignup(rec, req)

		res := rec.Result()
		if res.StatusCode != tc.expectedStatus {
			t.Fatalf("case %d failed: expected %d got %d", i, tc.expectedStatus, res.StatusCode)
		}
	}
}

func TestConfirmReset(t *testing.T) {
	s := initServiceOrPanic()

	cases := []struct {
		token          string
		expectedStatus int
	}{
		{"valid", http.StatusFound},
		{"", http.StatusNotFound},
		{"invalid", http.StatusNotFound},
	}

	for i, tc := range cases {
		req := httptest.NewRequest("POST", "/", nil)
		req.SetPathValue("token", tc.token)

		rec := httptest.NewRecorder()
		s.confirmReset(rec, req)

		res := rec.Result()
		if res.StatusCode != tc.expectedStatus {
			t.Fatalf("case %d failed: expected %d got %d", i, tc.expectedStatus, tc.expectedStatus)
		}
	}
}
