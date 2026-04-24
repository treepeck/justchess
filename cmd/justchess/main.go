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
	ar := db.NewSQLAuthRepo(pool)
	pr := db.NewSQLPlayerRepo(pool)
	gr := db.NewSQLGameRepo(pool)

	log.Print("Initializing services...")
	authService := auth.NewService(ar)
	if err = authService.ParseEmails("./_web/templates/emails/"); err != nil {
		log.Panic(err)
	}

	staticPages, err := web.ParsePages("./_web/templates/static/")
	if err != nil {
		log.Panic(err)
	}
	dynamicPages, err := web.ParsePages("./_web/templates/dynamic/")
	if err != nil {
		log.Panic(err)
	}
	webService := web.NewService(pr, gr, staticPages, dynamicPages)

	wsService := ws.NewService(gr, pr)
	go wsService.ListenEvents()

	// Register routes.
	mux := http.NewServeMux()
	wsService.RegisterRoutes(authService, mux)
	webService.RegisterRoutes(authService, mux)
	authService.RegisterRoutes(mux)

	log.Print("Starting server.")
	log.Panic(http.ListenAndServe(":3502", mux))
}
