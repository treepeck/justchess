package main

import (
	"bufio"
	"context"
	"justchess/pkg/auth"
	"justchess/pkg/db"
	"justchess/pkg/game"
	"justchess/pkg/player"
	"justchess/pkg/ws"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

func main() {
	logFile, err := os.OpenFile("./log.txt", os.O_RDWR|os.O_APPEND, os.ModeAppend)
	if err != nil {
		log.Fatalf("cannot open 'log.txt' file for writing: %v\n", err)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(logFile)

	loadEnv()
	log.Println("environment variables are loaded successfully")

	db.Open()
	defer db.Pool.Close()

	mux := setupMux()
	err = http.ListenAndServeTLS(":443", "cert.pem", "key.pem", mux)
	if err != nil {
		log.Printf("%v\n", err)
	}
}

func setupMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/auth/", allowCors(auth.Mux()))

	mux.Handle("/api/player/", allowCors(isAuthorized(player.Mux())))

	mux.Handle("/game/", allowCors(isAuthorized(game.Mux())))

	h := ws.NewHub()
	mux.HandleFunc("/hub", isAuthorizedWS(h.HandleNewConnection))

	mux.HandleFunc("/room", isAuthorizedWS(func(rw http.ResponseWriter, r *http.Request) {
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
	}))

	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		// To allow SharedBufferArray (for engine).
		rw.Header().Add("Cross-Origin-Opener-Policy", "same-origin")
		rw.Header().Add("Cross-Origin-Embedder-Policy", "require-corp")
		if len(r.URL.Path) < 5 {
			http.ServeFile(rw, r, "./frontend/index.html")
			return
		}

		mime := r.URL.Path[len(r.URL.Path)-3 : len(r.URL.Path)]
		if mime != "png" && mime != "css" && mime != ".js" && mime != "ico" && mime != "asm" {
			http.ServeFile(rw, r, "./frontend/index.html")
			return
		}
		http.FileServer(http.Dir("frontend")).ServeHTTP(rw, r)
	})

	return mux
}

// allowCors handles the Cross-Origin-Resource-Sharing.
func allowCors(next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Access-Control-Allow-Origin", os.Getenv("DOMAIN"))
		rw.Header().Add("Access-Control-Allow-Credentials", "true")
		rw.Header().Add("Access-Control-Allow-Headers", "Authorization")
		rw.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")

		// Handle CORS preflight request.
		if r.Method == "OPTIONS" {
			rw.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(rw, r)
	}
}

// isAuthorized accepts requests from every role!
// Protected endpoints must check the subject's role further.
// Cannot be used to authenticate the WebSocket connection, since it does not support request headers.
func isAuthorized(next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")

		if len(h) < 100 || h[:7] != "Bearer " {
			log.Printf("unauthorized request: %s\n", r.RemoteAddr)
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		cms, err := auth.DecodeToken(h[7:], "ACCESS_TOKEN_SECRET")
		if err != nil {
			log.Printf("unauthorized request: %s\n", r.RemoteAddr)
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), auth.Cms, cms)
		next.ServeHTTP(rw, r.WithContext(ctx))
	}
}

// isAuthorizedWS decodes the access token url parameter.
func isAuthorizedWS(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		encoded := r.URL.Query().Get("access")
		cms, err := auth.DecodeToken(encoded, "ACCESS_TOKEN_SECRET")
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), auth.Cms, cms)
		next.ServeHTTP(rw, r.WithContext(ctx))
	}
}

// loadEnv reads dev.env file from the root directory and sets environment
// variables for the current process.
// Accepted format for variable: KEY=VALUE
// Comments which begin with '#' and empty lines are skipped.
func loadEnv() {
	f, err := os.Open("dev.env")
	if err != nil {
		log.Fatalf("cannot read .env file: %v\n", err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())

		// Skip empty lines and comments.
		if len(line) < 2 || line[0] == '#' {
			continue
		}

		pair := strings.SplitN(line, "=", 2)
		// Skip malformed variable.
		if len(pair) != 2 {
			continue
		}

		os.Setenv(pair[0], pair[1])
	}

	if err := s.Err(); err != nil {
		log.Fatalf("error while reading .env: %v\n", err)
	}
}
