package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"justchess/pkg/user"

	"github.com/google/uuid"
)

// handleCreateGuest creates a new user and generates a pair of tokens for him.
// NOTE: the access token does not send back, the frontend must implicitly
// fetch the access token by triggering the /tokens endpoint.
func handleCreateGuest(rw http.ResponseWriter, r *http.Request) {
	u := user.NewUser(uuid.New())

	_, refresh, err := generatePair(u.Id)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	setRefreshTokenCookie(rw, refresh)

	err = json.NewEncoder(rw).Encode(u)
	if err != nil {
		log.Printf("%v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// handleRefreshTokens parses refresh token from the request cookie and
// generates a new pair if the token is valid.
func handleRefreshTokens(rw http.ResponseWriter, r *http.Request) {
	encoded, err := parseRefreshToken(r)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	id, err := decodeIdFromRefreshToken(encoded)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	access, refresh, err := generatePair(id)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	setRefreshTokenCookie(rw, refresh)
	rw.Header().Add("Content-Type", "text/plain")
	rw.Write([]byte(access))
}

// handleGetUserByRefreshToken parses the request cookie and fetches the user
// by the encoded in the token id.
func handleGetUserByRefreshToken(rw http.ResponseWriter, r *http.Request) {
	encoded, err := parseRefreshToken(r)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	id, err := decodeIdFromRefreshToken(encoded)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	// TODO: fetch registered users from the database by id.
	// To avoid creating a new guest each time the user refreshes the page
	// (for enabling game reconnection, etc.), send back the guest with an
	// old id.
	u := user.NewUser(id)

	err = json.NewEncoder(rw).Encode(u)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// setRefreshTokenCookie sets the Authorization cookie to the encoded refresh JWT.
func setRefreshTokenCookie(rw http.ResponseWriter, et string) {
	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    "Bearer " + et,
		Path:     "/",
		MaxAge:   2592000, // Token will last 30 days.
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(rw, &cookie)
}

// parseRefreshToken parses the refresh token from the request cookie.
func parseRefreshToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(cookie.Value, "Bearer ") {
		return "", err
	}
	et := strings.TrimPrefix(cookie.Value, "Bearer ")
	return et, nil
}

// decodeIdFromRefreshToken decodes the provided token and tries
// to parse the uuid from the token subject.
func decodeIdFromRefreshToken(et string) (uuid.UUID, error) {
	rt, err := DecodeToken(et, 2)
	if err != nil {
		return uuid.Nil, err
	}

	idStr, err := rt.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}
