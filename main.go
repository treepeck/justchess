package main

import (
	"chess-api/auth"
	"chess-api/db"
	"chess-api/middleware"
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
	fn := slog.String("func", "main")

	// load environment variables
	err := godotenv.Load()
	if err != nil {
		slog.Error(".env file cannot be load", fn, "err", err)
	}

	// connect to the database
	err = db.OpenDatabase()
	if err != nil {
		slog.Error("cannot open db", fn, "err", err)
	}
	slog.Info("Database connected successfully", fn)
	defer db.CloseDatabase()

	// create a middleware stack to send the
	// the request through the chain of middlewares
	middlewareStack := middleware.CreateStack(
		middleware.LogRequest,
		middleware.AllowCors,
	)

	// create a manager (basically same as router)
	// to handle websocket connections
	m := ws.NewManager()

	// load routes
	router := http.NewServeMux()
	router.Handle("/auth/", http.StripPrefix(
		"/auth",
		middlewareStack(auth.AuthRouter()),
	))
	router.HandleFunc("/ws", m.HandleConnection)

	// start server
	HOST := os.Getenv("SERVER_HOST")
	PORT := os.Getenv("SERVER_PORT")
	http.ListenAndServe(HOST+":"+PORT, router)
}
