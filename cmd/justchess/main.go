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

var (
	templates = os.DirFS(os.Getenv("TEMPLATES_DIR"))
	public    = os.DirFS(os.Getenv("PUBLIC_DIR"))
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	log.Print("Connecting to db.")
	pool, err := db.OpenDB(os.Getenv("MYSQL_URL"))
	if err != nil {
		log.Panic(err)
	}
	defer pool.Close()

	// Initialize database repository.
	repo := db.NewRepo(pool)
	log.Print("Successfully connected to db.")

	log.Print("Creating endpoints.")
	mux := http.NewServeMux()

	authService := auth.NewService(repo)
	authService.RegisterRoutes(mux)

	webService := web.NewService(templates, public, repo)
	webService.RegisterRoutes(mux)

	wsService := ws.NewService(repo)
	go wsService.EventBus()
	wsService.RegisterRoutes(mux)

	log.Print("Starting server.")
	log.Panic(http.ListenAndServe(":3502", mux))
}
