package main

import (
	"bufio"
	"context"
	"justchess/pkg/auth"
	"justchess/pkg/db"
	"justchess/pkg/ws"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	loadEnv()
	log.Println("environment variables are loaded successfully")

	db.Open()
	defer db.Pool.Close()

	mux := setupMux()
	err := http.ListenAndServe("localhost:3502", mux)
	if err != nil {
		log.Printf("%v\n", err)
	}
}

func setupMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /auth", allowCors(auth.RefreshHandler))
	mux.HandleFunc("GET /auth/guest", allowCors(auth.GuestHandler))
	mux.HandleFunc("GET /auth/verify", allowCors(auth.VerifyHandler))
	mux.HandleFunc("POST /auth/signup", allowCors(auth.SignUpHandler))
	mux.HandleFunc("POST /auth/signin", allowCors(auth.SignInHandler))
	mux.HandleFunc("POST /auth/reset", allowCors(auth.PasswordResetHandler))

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

	return mux
}

// allowCors handles the Cross-Origin-Resource-Sharing.
func allowCors(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000")
		rw.Header().Add("Access-Control-Allow-Credentials", "true")
		rw.Header().Add("Access-Control-Allow-Headers", "*")
		rw.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")

		// Handle CORS preflight request.
		if r.Method == "OPTIONS" {
			rw.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(rw, r)
	}
}

// isAuthorized accepts requests from every role.
// Protected endpoints must check the subject's role further.
// Cannot be used to authenticate the WebSocket connection, since it does not support request headers.
func isAuthorized(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")

		if len(h) < 100 || h[:7] != "Bearer " {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		subj, err := auth.DecodeToken(h[7:], "ACCESS_TOKEN_SECRET")
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), auth.Cms, subj)
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
