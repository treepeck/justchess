package middleware

import (
	"net/http"
	"os"
)

/*
AllowCORS handles the Cross-Origin-Resource-Sharing policy by allowing processing
only the requests that are coming from the trusted origin.
*/
func AllowCORS(next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		trustedDomain := os.Getenv("TRUSTED_DOMAIN")
		if r.Header.Get("Origin") != trustedDomain {
			http.Error(rw, "Origin domain is not trusted.", http.StatusForbidden)
			return
		}

		rw.Header().Add("Access-Control-Allow-Origin", trustedDomain)
		rw.Header().Add("Access-Control-Allow-Credentials", "true")
		rw.Header().Add("Access-Control-Allow-Headers", "Authorization")
		rw.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")

		// Handle CORS preflight request.
		if r.Method == "OPTIONS" {
			rw.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(rw, r)
	}
}
