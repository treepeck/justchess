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

	mux.HandleFunc("POST /auth/signup", allowCors(auth.SignUpHandler))
	mux.HandleFunc("GET /auth/verify", allowCors(auth.VerifyMailHandler))
	mux.HandleFunc("GET /auth/", allowCors(isAuthorized(auth.RefreshHandler)))
	mux.HandleFunc("POST /auth/reset", allowCors(auth.PasswordResetIssuer))
	mux.HandleFunc("POST /auth/reset-confirm", allowCors(auth.PasswordResetHandler))

	h := ws.NewHub()
	mux.HandleFunc("/hub", isAuthorized(h.HandleNewConnection))

	mux.HandleFunc("/room", isAuthorized(func(rw http.ResponseWriter, r *http.Request) {
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
		if r.URL.Host != "http://localhost:3000" {
			log.Printf("request from unknown host: %s\n", r.URL.Host)
			return
		}

		rw.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000")
		rw.Header().Add("Access-Control-Allow-Credentials", "true")
		rw.Header().Add("Access-Control-Allow-Headers", "origin, content-type, accept, authorization")
		rw.Header().Add("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS")

		// Handle CORS preflight request.
		if r.Method == "OPTIONS" {
			rw.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(rw, r)
	}
}

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

		ctx := context.WithValue(r.Context(), auth.Subj, subj)

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
