package main

import (
	"log"
	"net/http"
	"os"

	"justchess/internal/auth"
	"justchess/internal/db"
	"justchess/internal/web"
	"justchess/internal/ws"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	log.Print("Connecting to db...")
	pool, err := db.OpenDB(os.Getenv("MYSQL_URL"))
	if err != nil {
		log.Panic(err)
	}
	defer pool.Close()

	// Initialize database repository.
	repo := db.NewRepo(pool)
	log.Print("Successfully connected to db.")

	log.Print("Initializing services...")
	mux := http.NewServeMux()

	authService := auth.NewService(repo)
	authService.RegisterRoutes(mux)

	// Parse and store page templates.
	webService, err := web.InitService(repo)
	if err != nil {
		log.Panic(err)
	}
	webService.RegisterRoutes(mux)

	wsService := ws.NewService(repo)
	wsService.RegisterRoutes(mux)
	go wsService.ListenEvents()
	log.Print("Successfully initialized services.")

	log.Print("Starting server.")
	log.Panic(http.ListenAndServe(":3502", mux))
}
