package main

import (
	"log"
	"net/http"

	"justchess/pkg/auth"
	"justchess/pkg/middleware"
	"justchess/pkg/ws"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	mux := setupMux()
	err := http.ListenAndServe(":3502", mux)
	if err != nil {
		log.Printf("%v\n", err)
	}
}

func setupMux() *http.ServeMux {
	// Setup the chain of middlewares.
	authStack := middleware.CreateStack(
		middleware.LogRequest,
		middleware.AllowCors,
	)

	mux := http.NewServeMux()

	mux.Handle("/auth/", http.StripPrefix(
		"/auth",
		authStack(auth.AuthRouter()),
	))

	h := ws.NewHub()
	go h.EventPump()

	mux.HandleFunc("/ws", h.HandleNewConnection)
	return mux
}
