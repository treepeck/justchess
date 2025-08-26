package main

import (
	"justchess/internal/auth"
	"justchess/internal/core"
	"justchess/internal/db"
	"justchess/internal/player"
	"log"
	"net/http"
	"os"

	"github.com/rabbitmq/amqp091-go"

	"github.com/BelikovArtem/chego"
	"github.com/BelikovArtem/gatekeeper/pkg/env"
	"github.com/BelikovArtem/gatekeeper/pkg/mq"
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
	log.Printf("Successfully connected to RabbitMQ.")

	log.Print("Initializing services.")
	mux := http.NewServeMux()

	if err = player.InitPlayerService(pool, mux); err != nil {
		log.Panic(err)
	}

	if err = auth.InitAuthService(pool, mux); err != nil {
		log.Panic(err)
	}

	log.Print("Successfully initialized services.")

	// Initialize attack tables to be able to generate chess moves.
	chego.InitAttackTables()
	// Initialize Zobrist keys to be able to detect threefold repetitions.
	chego.InitZobristKeys()

	log.Print("Starting server.")
	c := core.NewCore(ch)

	// Run the goroutines which will run untill the program exits.
	go c.Handle()
	go mq.Consume(ch, "gate", c.Bus)

	// Set up router.
	http.ListenAndServe("localhost:3502", mux)
}
