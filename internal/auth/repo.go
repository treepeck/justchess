package auth

import "time"

type Session struct {
	Id        string
	PlayerId  string
	ExpiresAt time.Time
}

const (
	createStmt int = iota
	insertStmt
	selectStmt
	deleteStmt
)

const (
	createQuery = `CREATE TABLE IF NOT EXISTS session (
		id CHAR(8) PRIMARY KEY,
		player_id CHAR(8) NOT NULL UNIQUE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME AS (created_at + INTERVAL 24 HOUR) STORED,
		FOREIGN KEY (player_id) REFERENCES player(id)
	);`

	insertQuery = `INSERT INTO session (id, player_id) VALUES (?, ?);`
)
