package auth

import "net/http"

func AuthMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleGetUserByRefreshToken)
	mux.HandleFunc("PUT /", handleCreateGuest)
	mux.HandleFunc("GET /tokens", handleRefreshTokens)
	return mux
}
