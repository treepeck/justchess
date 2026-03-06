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
	pool, err := db.OpenDB(os.Getenv("DB_DSN"))
	if err != nil {
		log.Panic(err)
	}
	defer pool.Close()
	log.Print("Successfully connected to db.")

	// Initialize database repositories.
	pr := db.NewPlayerRepo(pool)
	gr := db.NewGameRepo(pool)

	log.Print("Initializing services...")
	authService := auth.NewService(pr)
	// Parse and store page templates.
	webService, err := web.InitService(pr, gr)
	if err != nil {
		log.Panic(err)
	}

	wsService := ws.NewService(pr, gr)
	go wsService.ListenEvents()
	log.Print("Successfully initialized services.")

	// Register routes.
	mux := http.NewServeMux()
	wsService.RegisterRoute(mux)
	webService.RegisterRoutes(mux)
	authService.RegisterRoutes(mux)

	log.Print("Starting server.")
	log.Panic(http.ListenAndServe(":3502", mux))
}
