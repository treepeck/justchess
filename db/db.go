package db

import (
	"database/sql"
	"fmt"
	"log"
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
	queryText := `
		CREATE TABLE IF NOT EXISTS users(
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			name VARCHAR(36) NOT NULL,
			password VARCHAR(36) NOT NULL,
			blitz_rating INT DEFAULT 400,
			rapid_rating INT DEFAULT 400,
			bullet_rating INT DEFAULT 400,
			games_count INT DEFAULT 0,
			likes INT DEFAULT 0,
			is_deleted BOOLEAN DEFAULT FALSE,
			registered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_visit TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = DB.Query(queryText)
	if err != nil {
		log.Fatalln(err)
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
