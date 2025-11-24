package db

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// OpenDB opens a database using a MySQL driver and verifies the specified
// connection url by calling a Ping method.  Sets the important connection
// parameters after a successful Ping.
func OpenDB(url string) (*sql.DB, error) {
	// Create a database pool.
	db, err := sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}

	// Since sql.Open does not connect to a db, validate a url with db.Ping
	// before executing any queries.
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Set connection parameters.
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, err
}
