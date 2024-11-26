package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

// CreateStack creates a middleware stack.
// Each request will be send through the stack.
func CreateStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			next = x(next)
		}
		return next
	}
}
