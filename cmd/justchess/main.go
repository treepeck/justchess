package main

import (
	"justchess/internal/auth"
	"justchess/internal/db"
	"justchess/internal/env"
	"justchess/internal/tmpl"
	"justchess/internal/ws"
	"log"
	"net/http"

	"github.com/BelikovArtem/chego"
)

func main() {
	mux := setupMux()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	// Initialize attack tables.
	chego.InitAttackTables()

	env.Load(".env")

	db.Open()
	defer db.Close()
	db.ApplySchema()

	http.ListenAndServe("localhost:3502", mux)
}

func setupMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/auth/", auth.Mux())

	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		tmpl.Exec(rw, "home.html")
	})

	mux.HandleFunc("/signup", func(rw http.ResponseWriter, r *http.Request) {
		tmpl.Exec(rw, "signup.html")
	})

	// TODO: secure file server against creation of symbolic links.
	mux.Handle("/static/", http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static")),
	))

	h := ws.NewHub()

	mux.HandleFunc("GET /websocket", auth.IsAuthorized(func(rw http.ResponseWriter,
		r *http.Request) {
		ws.HandleNewConnection(h, rw, r)
	}))

	return mux
}
