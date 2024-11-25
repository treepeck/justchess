package main

import (
	"chess-api/auth"
	"chess-api/db"
	"chess-api/middleware"
	"chess-api/user"
	"chess-api/ws"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// set up logger
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().Format("01/02/2006 15:04:05"))
			}
			return a
		},
	})
	slog.SetDefault(slog.New(h))
	// load environment variables
	err := godotenv.Load()
	if err != nil {
		slog.Error(".env file cannot be load", "err", err)
	}
	// connect to the database
	err = db.OpenDatabase("./db/schema.sql")
	if err != nil {
		return
	}
	slog.Info("Database connected successfully")
	defer db.CloseDatabase()
	// create a middleware stack to send the
	// the request through the chain of middlewares
	middlewareStack := middleware.CreateStack(
		middleware.LogRequest,
		middleware.AllowCors,
		middleware.IsAuthorized,
	)
	// instantiate manager (basically same as router)
	// to handle websocket connections
	m := ws.NewManager()
	// load routes
	router := http.NewServeMux()
	router.Handle("/auth/", http.StripPrefix(
		"/auth",
		middleware.LogRequest(middleware.AllowCors(auth.AuthRouter())),
	))
	router.Handle("/user/", http.StripPrefix(
		"/user",
		middlewareStack(user.UserRouter()),
	))
	router.HandleFunc("/ws", m.HandleConnection)
	// start server
	HOST := os.Getenv("SERVER_HOST")
	PORT := os.Getenv("SERVER_PORT")
	http.ListenAndServe(HOST+":"+PORT, router)
}
