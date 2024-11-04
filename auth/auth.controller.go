package auth

import (
	"chess-api/jwt_auth"
	"chess-api/models/user"
	"chess-api/repository"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

// handleGuest generates a new pair of tokens and sends them back to the client,
// so that client can play but can not have access to rating system.
func handleGuest(rw http.ResponseWriter, r *http.Request) {
	id := uuid.New()
	at, rt, err := jwt_auth.GeneratePair(jwt_auth.Subject{
		Id: id,
		R:  jwt_auth.Guest,
	})
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	setRefreshTokenCookie(rw, rt)
	// send access token back to the client
	rw.Header().Add("Content-Type", "text/plain")
	rw.Write([]byte(at))
}

// handleGetTokens parses the encoded refresh token from a cookie and tries to
// decode it. If success, it generates a new pair of tokens (refresh and
// access tokens) and sends them back to the client to keep the user authorized.
func handleGetTokens(rw http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(cookie.Value, "Bearer ") {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	et := strings.TrimPrefix(cookie.Value, "Bearer")

	rt, err := jwt_auth.DecodeToken(et, "REFRESH_TOKEN_SECRET")
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	es, err := rt.Claims.GetSubject()
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	var s jwt_auth.Subject
	err = json.Unmarshal([]byte(es), &s)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	if s.R == jwt_auth.Guest {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	at, _rt, err := jwt_auth.GeneratePair(s)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	setRefreshTokenCookie(rw, _rt)

	rw.Header().Add("Content-Type", "text/plain")
	rw.Write([]byte(at))
}

// handleGetUserByAccessToken returns the user data
func handleGetUserByAccessToken(rw http.ResponseWriter, r *http.Request) {
	// parse access token from the Authorization header
	h := r.Header.Get("Authorization")
	// encoded JWT string must begin with the Bearer prefix
	if !strings.HasPrefix(h, "Bearer ") {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	// parse encoded token
	et := strings.TrimPrefix(h, "Bearer ")

	at, err := jwt_auth.DecodeToken(et, "ACCESS_TOKEN_SECRET")
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	// try to get user info from the decoded access token
	es, err := at.Claims.GetSubject()
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	var s jwt_auth.Subject
	err = json.Unmarshal([]byte(es), &s)
	if err != nil {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	// return mock guest data
	var u user.U
	if s.R == jwt_auth.Guest {
		u = *user.NewUser(s.Id)
	} else {
		_u := repository.FindUserById(s.Id)
		if _u == nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
		u = *_u
	}
	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(u)
}

// setRefreshTokenCookie sets the Authorization cookie to the encoded refresh JWT.
func setRefreshTokenCookie(rw http.ResponseWriter, et string) {
	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    "Bearer " + et,
		Path:     "/",
		Domain:   os.Getenv("SERVER_HOST"),
		MaxAge:   2592000, // 30 days in seconds
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(rw, &cookie)
}
