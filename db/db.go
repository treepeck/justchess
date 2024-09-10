package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func OpenDatabase() error {
	// format a connection string
	connectStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	var err error
	DB, err = sql.Open("postgres", connectStr)
	if err != nil {
		return err
	}

	// create tables
	schema, err := os.ReadFile("./db/schema.sql")
	if err != nil {
		slog.Error("database schema not found", "err", err)
		return err
	}

	_, err = DB.Query(string(schema))
	if err != nil {
		slog.Error("query cannot be executed", "err", err)
	}

	return err
}

func CloseDatabase() error {
	if DB != nil {
		err := DB.Close()
		return err
	}
	return fmt.Errorf("database isn`t connected")
}
