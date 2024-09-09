package middleware

import (
	"net/http"
	"os"
)

func AllowCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		clientDomain := os.Getenv("CLIENT_DOMAIN")
		rw.Header().Add("Access-Control-Allow-Origin", clientDomain)
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
