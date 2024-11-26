package middleware

import (
	"net/http"
)

// AllowCors
func AllowCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000")
		rw.Header().Add("Access-Control-Allow-Credentials", "true")
		rw.Header().Add("Access-Control-Allow-Headers", "origin, content-type, accept, 	authorization")

		// handle CORS preflight request
		if r.Method == "OPTIONS" {
			rw.WriteHeader(200)
			return
		}
		next.ServeHTTP(rw, r)
	})
}
