package main

import (
	"log"
	"net/http"
	"os"

	"justchess/pkg/auth"
	"justchess/pkg/middleware"
	"justchess/pkg/ws"
)

func main() {
	if len(os.Args) < 3 {
		log.Printf("Provide the required arguments to run the program.\n1 - ACCESS_TOKEN_SECRET;\n2 - REFRESH_TOKEN_SECRET.\n")
		return
	}

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
		middleware.AllowCors,
		middleware.LogRequest,
	)

	mux := http.NewServeMux()
	mux.Handle("/auth/", http.StripPrefix(
		"/auth",
		authStack(auth.AuthMux()),
	))

	h := ws.NewHub()
	go h.EventPump()

	mux.HandleFunc("/ws", h.HandleNewConnection)
	return mux
}
