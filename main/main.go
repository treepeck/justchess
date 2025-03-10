package main

import (
	"justchess/pkg/auth"
	"justchess/pkg/ws"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
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
	mux := http.NewServeMux()

	mux.Handle("/auth/", http.StripPrefix(
		"/auth",
		AllowCors(auth.AuthMux()),
	))

	h := ws.NewHub()

	mux.HandleFunc("/hub", h.HandleNewConnection)

	mux.HandleFunc("/room", func(rw http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.URL.Query().Get("id"))
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		room := h.GetRoomById(id)
		if room == nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		room.HandleNewConnection(rw, r)
	})

	return mux
}

// AllowCors handles the Cross-Origin-Resource-Sharing.
func AllowCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000")
		rw.Header().Add("Access-Control-Allow-Credentials", "true")
		rw.Header().Add("Access-Control-Allow-Headers", "origin, content-type, accept, authorization")
		rw.Header().Add("Access-Control-Allow-Methods", "GET,PUT,OPTIONS")

		// Handle CORS preflight request.
		if r.Method == "OPTIONS" {
			rw.WriteHeader(200)
			return
		}
		next.ServeHTTP(rw, r)
	})
}
