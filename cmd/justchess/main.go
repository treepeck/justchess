package main

import (
	"log"
	"net/http"
	"os"

	"justchess/internal/auth"
	"justchess/internal/core"
	"justchess/internal/db"
	"justchess/internal/tmpl"

	"github.com/treepeck/chego"
	"github.com/treepeck/gatekeeper/pkg/env"
	"github.com/treepeck/gatekeeper/pkg/mq"

	"github.com/rabbitmq/amqp091-go"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	log.Print("Loading environment variables.")
	if err := env.Load(".env"); err != nil {
		log.Panic(err)
	}
	log.Print("Successfully loaded environment variables.")

	log.Print("Connecting to db.")
	pool, err := db.OpenDB(os.Getenv("MYSQL_URL"))
	if err != nil {
		log.Panic(err)
	}
	defer pool.Close()
	log.Print("Successfully connected to db.")

	log.Print("Initializing tables.")
	repo := db.NewRepo(pool)
	if err = repo.CreatePlayer(); err != nil {
		log.Panic(err)
	}

	if err = repo.CreateSession(); err != nil {
		log.Panic(err)
	}
	// TODO: initialize game table.
	log.Print("Tables are successfully initialized.")

	log.Print("Connecting to RabbitMQ.")
	conn, err := amqp091.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	// Open an AMQP channel.
	ch, err := conn.Channel()
	if err != nil {
		log.Panic(err)
	}
	defer ch.Close()

	// Put the channel into a confirm mode.
	if err = ch.Confirm(false); err != nil {
		log.Panic(err)
	}
	log.Print("Successfully connected to RabbitMQ.")

	log.Print("Creating endpoints.")
	mux := http.NewServeMux()

	authService := auth.NewService(repo)
	mux.HandleFunc("POST /auth/signup", authService.HandleSignup)
	mux.HandleFunc("POST /auth/signin", authService.HandleSignin)
	mux.HandleFunc("POST /auth/verify", authService.HandleVerify)

	// Serve static frontend files.
	fs := http.Dir("./static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(fs)))

	// Serve HTML templates.
	mux.HandleFunc("/", tmpl.Exec)

	// Initialize attack tables to be able to generate chess moves.
	chego.InitAttackTables()
	// Initialize Zobrist keys to be able to detect threefold repetitions.
	chego.InitZobristKeys()

	log.Print("Starting server.")
	c := core.NewCore(ch, pool)

	// Run the goroutines which will run untill the program exits.
	go c.Run()
	go mq.Consume(ch, "gate", c.EventBus)

	log.Panic(http.ListenAndServe("localhost:3502", mux))
}
