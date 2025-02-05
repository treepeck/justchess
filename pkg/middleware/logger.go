package middleware

import (
	"log"
	"net/http"
)

// LogRequest logs incomming http requests.
// It only skips CORS preflight requests with the OPTIONS method.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			next.ServeHTTP(rw, r)
			return
		}
		log.Printf("request: %v %v\n", r.Method, r.URL.String())
		next.ServeHTTP(rw, r)
	})
}
