package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var Pool *sql.DB

func Open() {
	var err error
	Pool, err = sql.Open("postgres", os.Getenv("DB_CONNSTRING"))

	if err != nil {
		log.Fatalf("cannot open database: %v\n", err)
	}

	if err = Pool.Ping(); err != nil {
		log.Fatalf("connection not established: %v\n", err)
	}
}
