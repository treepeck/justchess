package main

import (
	"chess-api/auth"
	"chess-api/db"
	"chess-api/ws"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

const (
	HOST = "localhost"
	PORT = "3502"
)

func main() {
	// load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}

	// connect to the database
	err = db.OpenDatabase()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("main: Database connected successfully")
	defer db.CloseDatabase()

	// create a manager (basically same as router)
	// to handle websocket connections
	m := ws.NewManager()

	// load routes
	router := http.NewServeMux()
	router.Handle("/auth/", http.StripPrefix(
		"/auth",
		auth.AuthRouter(),
	))
	router.HandleFunc("/ws", m.HandleConnection)

	// start server
	log.Fatalln(http.ListenAndServe(HOST+":"+PORT, router))
}
