package main

import (
	"log"
	"net/http"

	"justchess/pkg/auth"
	"justchess/pkg/middleware"
	"justchess/pkg/ws"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// Load environment variables.
	// _, err := os.ReadFile("../.env")
	// if err != nil {
	// 	log.Printf("%v\n", err)
	// 	return
	// }
	// Setup routes.
	router := setupRouter()
	err := http.ListenAndServe(":3502", router)
	if err != nil {
		log.Printf("%v\n", err)
	}
}

func setupRouter() *http.ServeMux {
	// Setup the chain of middlewares.
	authStack := middleware.CreateStack(
		middleware.LogRequest,
		middleware.AllowCors,
	)
	router := http.NewServeMux()
	router.Handle("/auth/", http.StripPrefix(
		"/auth",
		authStack(auth.AuthRouter()),
	))
	// Instantiate manager to handle ws connections.
	m := ws.NewManager()
	go m.Run()
	router.HandleFunc("/ws", m.HandleNewConnection)
	return router
}
