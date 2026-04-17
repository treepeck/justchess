package security

import "net/http"

// Headers is the middleware that writes recommended security headers for
// each incomming request.
func Headers(next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Security-Policy", "default-src 'self'; script-src 'self' 'wasm-unsafe-eval'")
		rw.Header().Add("Strict-Transport-Security", "max-age=63072000")
		rw.Header().Add("X-Content-Type-Options", "nosniff")
		rw.Header().Add("X-Robots_Tag", "index, follow")
		rw.Header().Add("X-Frame-Options", "DENY")

		next.ServeHTTP(rw, r)
	}
}
