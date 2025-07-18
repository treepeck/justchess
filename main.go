package main

import (
	"justchess/pkg/auth"
	"justchess/pkg/db"
	"justchess/pkg/ws"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/BelikovArtem/chego/movegen"
)

func main() {
	mux := setupMux()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	// Initialize attack tables.
	movegen.InitAttackTables()

	loadEnv()
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

// loadEnv reads .env file from the root directory and sets environment
// variables for the current process.
// Accepted format for a variable: KEY=VALUE
// Comments which begin with '#' and empty lines are ignored.
func loadEnv() {
	env, err := os.ReadFile(".env")
	if err != nil {
		log.Fatalf(".env file cannot be read %v", err)
	}

	for line := range strings.SplitSeq(string(env), "\n") {
		// Skip empty lines and comments.
		if len(line) < 3 || line[0] == '#' {
			continue
		}

		pair := strings.SplitN(line, "=", 2)
		// Skip malformed variable.
		if len(pair) != 2 {
			continue
		}

		os.Setenv(pair[0], pair[1])
	}
}
