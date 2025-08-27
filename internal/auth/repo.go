package auth

import (
	"database/sql"
	"time"
)

type Session struct {
	Id        string
	PlayerId  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

const (
	createQuery = `CREATE TABLE IF NOT EXISTS session (
		id CHAR(32) PRIMARY KEY,
		player_id CHAR(12) NOT NULL UNIQUE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME AS (created_at + INTERVAL 24 HOUR) STORED,
		FOREIGN KEY (player_id) REFERENCES player(id)
	)`

	insertQuery = `INSERT INTO session (id, player_id) VALUES (?, ?)`

	selectByIdQuery = `SELECT * FROM session WHERE id = ?`

	deleteExpiredQuery = `DELETE FROM session WHERE expires_at < NOW()`
)

/*
Insert inserts a single record into the session table.
*/
func Insert(pool *sql.DB, id, playerId string) error {
	_, err := pool.Exec(insertQuery, id, playerId)
	return err
}

/*
SelectById selects a single record with the same id as provided from the session
table.

NOTE: It may return an expired session.  It's a caller's responsibility to
delete all expired sessions before calling this function.
*/
func SelectById(pool *sql.DB, id string) (Session, error) {
	row := pool.QueryRow(selectByIdQuery, id)

	var s Session
	err := row.Scan(&s.Id, &s.PlayerId, &s.CreatedAt, &s.ExpiresAt)
	return s, err
}

/*
DeleteExpired delets all expired records from the session table.
*/
func DeleteExpired(pool *sql.DB) error {
	_, err := pool.Exec(deleteExpiredQuery)
	return err
}
