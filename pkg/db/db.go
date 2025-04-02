package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var Pool *sql.DB

// Open opens the db connection and executes the schema.sql file.
func Open() {
	var err error
	Pool, err = sql.Open("postgres", os.Getenv("DB_CONNSTRING"))

	if err != nil {
		log.Fatalf("cannot open database: %v\n", err)
	}

	if err = Pool.Ping(); err != nil {
		log.Fatalf("connection not established: %v\n", err)
	}

	schema, err := os.ReadFile("./pkg/db/schema.sql")
	if err != nil {
		log.Fatalf("cannot read schema.sql: %v\n", err)
	}

	if _, err := Pool.Query(string(schema)); err != nil {
		log.Fatalf("cannot apply schema.sql: %v\n", err)
	}
}
