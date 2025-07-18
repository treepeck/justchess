package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// Database connection pool.
var pool *sql.DB

// Open opens the db connection.
func Open() {
	var err error

	dsn := os.Getenv("DSN")
	pool, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if err = pool.Ping(); err != nil {
		log.Fatalf("%v", err)
	}

	pool.SetConnMaxLifetime(0)
	pool.SetMaxIdleConns(3)
	pool.SetMaxOpenConns(3)

}

// ApplySchma executes queries from the schema.sql file
func ApplySchema() {
	schema, err := os.ReadFile("./pkg/db/schema.sql")
	if err != nil {
		pool.Close()
		log.Fatalf("%v", err)
	}

	if _, err = pool.Query(string(schema)); err != nil {
		log.Fatalf("%v", err)
	}
}

// Close closes the db connection.
func Close() {
	pool.Close()
}
