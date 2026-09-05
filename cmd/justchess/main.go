package main

import (
	"log"
	"net/http"
	"os"

	"justchess/internal/api"
	"justchess/internal/auth"
	"justchess/internal/db"
	"justchess/internal/game"
	"justchess/internal/web"
	"justchess/internal/ws"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	cookieKey, err := auth.ParseCookieKey(os.Getenv("COOKIE_KEY"))
	if err != nil {
		log.Panic(err)
	}

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

	gs := game.NewStorage(gr)
	go gs.Listen()

	log.Print("Initializing services...")
	authService := auth.NewService(cookieKey, ar)
	if err = authService.ParseEmails("./internal/auth/templates/"); err != nil {
		log.Panic(err)
	}

	apiService := api.NewService(gr, pr)

	webService, err := web.InitService(gr, pr, "./_web/")
	if err != nil {
		log.Panic(err)
	}

	wsService := ws.NewService()
	// go wsService.Listen()

	// Register routes.
	mux := http.NewServeMux()
	authService.RegisterRoutes(mux)
	apiService.RegisterRoutes(authService, mux)
	wsService.RegisterRoutes(mux)
	webService.RegisterRoutes(authService, mux)

	log.Print("Starting server.")
	log.Panic(http.ListenAndServe(":3502", mux))
}
