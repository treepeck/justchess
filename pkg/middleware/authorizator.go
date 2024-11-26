package middleware

import (
	"context"
	"justchess/pkg/auth"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// A unique key to get and set userId from ruests
type authKey struct {
	id string
}

var IdKey = authKey{"middleware.auth.user"}

// IsAuthorized decodes the access JWT from the Authorization header
// and passes it to the next handler as a request context.
func IsAuthorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")

		if !strings.HasPrefix(h, "Bearer ") {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		et := strings.TrimPrefix(h, "Bearer ")

		at, err := auth.DecodeToken(et, "ATS")
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		idStr, err := at.Claims.GetSubject()
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Pass decoded user id with the request.
		ctx := context.WithValue(r.Context(), IdKey, id)
		uR := r.WithContext(ctx)
		// User is authorized, continue processing the request.
		next.ServeHTTP(rw, uR)
	})
}
