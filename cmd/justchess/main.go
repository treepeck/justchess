package main

import (
	"log"
	"net/http"
	"os"

	"github.com/treepeck/justchess/internal/api"
	"github.com/treepeck/justchess/internal/transport"
	"github.com/treepeck/justchess/internal/web"
	"github.com/treepeck/justchess/pkg/auth"
	"github.com/treepeck/justchess/pkg/db"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	log.Print("Parsing COOKIE_KEY environment variable...")
	cookieKey, err := auth.ParseCookieKey(os.Getenv("COOKIE_KEY"))
	if err != nil {
		log.Panic(err)
	}
	log.Print("Successfully parsed COOKIE_KEY.")

	log.Print("Connecting to db...")
	pool, err := db.OpenDB(os.Getenv("DB_DSN"))
	if err != nil {
		log.Panic(err)
	}
	defer pool.Close()
	log.Print("Successfully connected to db.")

	// Initialize database repositories.
	log.Print("Initializing database repositories...")
	ar := db.NewSQLAuthRepo(pool)
	pr := db.NewSQLPlayerRepo(pool)
	gr := db.NewSQLGameRepo(pool)
	log.Print("Successfully initialized database repositories.")

	log.Print("Initializing services...")
	authService := auth.NewService(cookieKey, ar)
	if err = authService.ParseEmails("./pkg/auth/templates/"); err != nil {
		log.Panic(err)
	}

	apiService := api.NewService(gr, pr)

	webService, err := web.InitService(gr, pr, "./_web/")
	if err != nil {
		log.Panic(err)
	}

	transport.InitService()
	/*
		if err != nil {
				log.Panic(err)
			}
	*/
	log.Print("Successfully initialized services.")

	// Register routes.
	log.Print("Registering HTTP endpoints...")
	mux := http.NewServeMux()
	authService.RegisterRoutes(mux)
	apiService.RegisterRoutes(authService, mux)
	webService.RegisterRoutes(authService, mux)
	log.Print("Successfully registered HTTP endpoints.")

	log.Print("Starting server.")
	log.Panic(http.ListenAndServe(":3502", mux))
}
