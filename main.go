package main

import (
	"justchess/pkg/auth"
	"justchess/pkg/db"
	"justchess/pkg/env"
	"justchess/pkg/ws"
	"log"
	"net/http"

	"github.com/BelikovArtem/chego/movegen"
)

func main() {
	mux := setupMux()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	// Initialize attack tables.
	movegen.InitAttackTables()

	env.Load(".env")
	db.Open()
	defer db.Close()
	db.ApplySchema()

	http.ListenAndServe("localhost:3502", mux)
}

func setupMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/auth/", auth.Mux())

	mux.Handle("/", http.FileServer(http.Dir("./frontend/page/home/")))
	mux.Handle("/play/", http.StripPrefix(
		"/play/",
		http.FileServer(http.Dir("./frontend/page/play")),
	))
	mux.Handle("/js/", http.StripPrefix(
		"/js/",
		http.FileServer(http.Dir("./frontend/js")),
	))

	h := ws.NewHub()

	mux.HandleFunc("GET /websocket", func(rw http.ResponseWriter, r *http.Request) {
		ws.HandleNewConnection(h, rw, r)
	})

	return mux
}
