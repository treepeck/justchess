package main

import (
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

	http.ListenAndServe("localhost:3502", mux)
}

func setupMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./frontend/page/home/")))
	mux.Handle("/play/", http.StripPrefix(
		"/play/",
		http.FileServer(http.Dir("./frontend/page/play")),
	))

	room := ws.NewRoom()
	go room.EventRoutine()

	mux.HandleFunc("/websocket", func(rw http.ResponseWriter, r *http.Request) {
		ws.HandleNewConnection(room, rw, r)
	})

	return mux
}
