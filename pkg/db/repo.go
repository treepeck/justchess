package db

import "database/sql"

// SelectSessionById selects a single row by the provided unique id.
// Session table has id as primary key to be selected by it.
func SelectSessionById(id string) *sql.Row {
	query := "SELECT user_id, expires_at FROM session WHERE id = $1;"
	return pool.QueryRow(query, id)
}

// InsertSessions inserts a new session and returns the session id.
func InsertSession() *sql.Row {
	query := "INSERT INTO session DEFAULT VALUES RETURNING id, user_id;"
	return pool.QueryRow(query)
}

// DeleteSession delets session with the provided id.
func DeleteSession(sid string) error {
	query := "DELETE FROM session WHERE id = $1;"
	_, err := pool.Exec(query, sid)
	return err
}
