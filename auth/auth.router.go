package auth

import "net/http"

func AuthRouter() (router *http.ServeMux) {
	router = http.NewServeMux()

	router.HandleFunc("GET /me", handleGetUserByAccessToken)
	router.HandleFunc("GET /guest", handleGuest)
	router.HandleFunc("GET /tokens", handleGetTokens)
	return
}
