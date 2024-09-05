package auth

import "net/http"

func AuthRouter() (router *http.ServeMux) {
	router = http.NewServeMux()

	router.HandleFunc("POST /guest", handleGuest)
	return
}
