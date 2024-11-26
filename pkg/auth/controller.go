package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"justchess/pkg/models/user"
	"justchess/pkg/repository"

	"github.com/google/uuid"
)

// handleCreateGuest creates a new guest and generates a pair of tokens for him.
func handleCreateGuest(rw http.ResponseWriter, r *http.Request) {
	g := user.NewGuest()
	at, rt, err := generatePair(g.Id)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	g.AccessToken = at
	setRefreshTokenCookie(rw, rt)

	err = json.NewEncoder(rw).Encode(g)
	if err != nil {
		slog.Error("Guest cannot be encoded.", "err", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// handleRefreshTokens parses refresh token from the request cookie and
// generates a new pair if the token is valid.
func handleRefreshTokens(rw http.ResponseWriter, r *http.Request) {
	et, err := parseRefreshToken(r)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	id, err := decodeIdFromRefreshToken(et)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	at, nrt, err := generatePair(id)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	setRefreshTokenCookie(rw, nrt)
	rw.Header().Add("Content-Type", "text/plain")
	rw.Write([]byte(at))
}

// handleGetUserByRefreshToken parses the request cookie and fetches the user
// by the encoded in token id.
func handleGetUserByRefreshToken(rw http.ResponseWriter, r *http.Request) {
	et, err := parseRefreshToken(r)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	id, err := decodeIdFromRefreshToken(et)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Generate a new pair for authenticated user to keep them signed in.
	at, nrt, err := generatePair(id)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	setRefreshTokenCookie(rw, nrt)

	u := repository.FindUserById(id)
	// In this case the user is missing in a db, but has a valid JWT,
	// so we deal with a guest.
	if u == nil {
		// To avoid creating a new guest each time the user refreshes the page
		// (for enabling room reconnection, etc.), send back the guest with an
		// old id.
		g := user.Guest{
			Id:          id,
			Name:        "Guest-" + id.String()[0:8],
			AccessToken: at,
		}
		err := json.NewEncoder(rw).Encode(g)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		err := json.NewEncoder(rw).Encode(u)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
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
	rt, err := DecodeToken(et, "RTS")
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
