package middleware

import (
	"log/slog"
	"net/http"
)

// LogRequest logs each incomming http request.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			next.ServeHTTP(rw, r)
			return
		}

		slog.Info("Request", slog.String("method", r.Method),
			slog.String("url", r.URL.String()))

		next.ServeHTTP(rw, r)
	})
}
