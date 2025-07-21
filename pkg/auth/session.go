package auth

import (
	"context"
	"justchess/pkg/db"
	"log"
	"net/http"
)

type CtxKey string

const PidKey CtxKey = "pid"

// IsAuthorized middleware decodes and validates the session from the Authorization cookie.
func IsAuthorized(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if len(r.Cookies()) != 1 || r.Cookies()[0].Name != "Authorization" {
			http.Error(rw, "Missing cookie", http.StatusUnauthorized)
			return
		}

		sid := r.Cookies()[0].Value

		err := db.DeleteExpiredSessions()
		if err != nil {
			log.Printf("ERROR: cannot delete expired sessions %v", err)
			http.Error(rw, "Cennot delete expired sessions", http.StatusInternalServerError)
			return
		}

		pid, err := db.SelectPlayerIdBySessionId(sid)
		if err != nil {
			http.Error(rw, "Session not active", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), PidKey, pid)
		next.ServeHTTP(rw, r.WithContext(ctx))
	}
}

func setAutorizationCookie(rw http.ResponseWriter, sid string) {
	c := http.Cookie{
		Name:     "Authorization",
		Value:    sid,
		Path:     "/",
		MaxAge:   86400, // Session will last 1 day.
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(rw, &c)
}
