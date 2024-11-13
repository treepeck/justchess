package middleware

import (
	"chess-api/jwt_auth"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// A unique key to get and set userId from ruests
type authKey struct {
	id string
}

var IdKey = authKey{"middleware.auth.user"}

// IsAuthorized checks the ruest headers and tries to decode the access JWT from the Authorization header.
func IsAuthorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/guest" || r.URL.String() == "/tokens" {
			next.ServeHTTP(rw, r)
			return
		}

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

		// pass user id through all routing stack
		ctx := context.WithValue(r.Context(), IdKey, s)
		uR := r.WithContext(ctx)
		// user is authorized, continue processing the request
		next.ServeHTTP(rw, uR)
	})
}
