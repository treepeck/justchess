package user

import "net/http"

func UserRouter() (router *http.ServeMux) {
	router = http.NewServeMux()

	router.HandleFunc("GET /id/", handleGetUserById)
	return
}
