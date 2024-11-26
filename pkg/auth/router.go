package auth

import "net/http"

func AuthRouter() (router *http.ServeMux) {
	router = http.NewServeMux()

	router.HandleFunc("GET /guest", handleCreateGuest)
	router.HandleFunc("GET /tokens", handleRefreshTokens)
	router.HandleFunc("GET /me", handleGetUserByRefreshToken)
	return
}
