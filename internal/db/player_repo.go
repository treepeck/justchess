package db

import (
	"database/sql"
	"time"
)

// The following block of constants declares SQL queries used to access and
// modify the player-related tables.
const (
	insertPlayer = `
	INSERT INTO player (
		id,
		name,
		email,
		password_hash
	)
	VALUES (?, ?, ?, ?)`

	selectPlayerById = `SELECT * FROM player WHERE id = ?`

	selectPlayerByEmail = `SELECT * FROM player WHERE email = ?`

	selectPlayerBySessionId = `
	SELECT
		p.*
	FROM player p
	INNER JOIN session s
	ON p.id = s.player_id
	WHERE s.id = ? AND s.expires_at > NOW()`

	insertSession = `
	INSERT INTO session (
		id,
		player_id
	)
	VALUES (?, ?)`

	selectSessionById = `
	SELECT * FROM session
	WHERE
		id = ?
	AND
		expires_at > NOW()`

	selectSessionsByPlayerId = `
	SELECT * FROM session
	WHERE
		player_id = ?
	AND
	expires_at > NOW()`

	deleteSessionById = `DELETE FROM session WHERE id = ?`
)

// Player represents a registered player.  Sensitive data, such as password hash
// and email will not be encoded into a JSON.
type Player struct {
	PasswordHash     []byte
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Id               string
	Name             string
	Email            string
	Rating           float64
	RatingDeviation  float64
	RatingVolatility float64
}

// Session is an authorization token for a player.  Each protected endpoint
// expects the Auth cookie to contain valid and not expired session.
type Session struct {
	CreatedAt time.Time
	ExpiresAt time.Time
	Id        string
	PlayerId  string
}

// PlayerRepo wraps the database connection pool and provides methods to access
// and modify the player table.
type PlayerRepo struct {
	pool *sql.DB
}

func NewPlayerRepo(p *sql.DB) PlayerRepo { return PlayerRepo{pool: p} }

// Insert inserts a single record into the player table, using the provided
// credentials.
func (r PlayerRepo) Insert(id, name, email string, passwordHash []byte) error {
	_, err := r.pool.Exec(insertPlayer, id, name, email, passwordHash)
	return err
}

// SelectById selects a single record with the same id as provided from the
// player table.
func (r PlayerRepo) SelectById(id string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerById, id)

	var p Player
	return p, row.Scan(&p.Id, &p.Name, &p.Rating, &p.RatingDeviation,
		&p.RatingVolatility, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
}

// SelectByEmail selects a single record with the same email as provided
// from the player table.
func (r PlayerRepo) SelectByEmail(email string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerByEmail, email)

	var p Player
	return p, row.Scan(&p.Id, &p.Name, &p.Rating, &p.RatingDeviation,
		&p.RatingVolatility, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
}

// SelectBySessionId selects a single player with the specified session_id.
// Expired sessions are omitted.
func (r PlayerRepo) SelectBySessionId(id string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerBySessionId, id)

	var p Player
	return p, row.Scan(&p.Id, &p.Name, &p.Rating, &p.RatingDeviation,
		&p.RatingVolatility, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
}

// InsertSession inserts a single record into the session table.
func (r PlayerRepo) InsertSession(id, playerId string) error {
	_, err := r.pool.Exec(insertSession, id, playerId)
	return err
}

// SelectByPlayerId selects a single record with the specified player_id
// from the session table.  Expired sessions are omitted.
func (r PlayerRepo) SelectSessionById(id string) (Session, error) {
	row := r.pool.QueryRow(selectSessionById, id)

	var s Session
	return s, row.Scan(&s.Id, &s.PlayerId, &s.CreatedAt, &s.ExpiresAt)
}

// SelectByPlayerId selects multiple records with the specified player_id
// from the session table.  Expired sessions are omitted.
func (r PlayerRepo) SelectSessionByPlayerId(playerId string) ([]Session, error) {
	rows, err := r.pool.Query(selectSessionsByPlayerId, playerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session

	for rows.Next() {
		var s Session
		if err = rows.Scan(&s.Id, &s.PlayerId, &s.CreatedAt, &s.ExpiresAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// DeleteSessionById delets single record with the same id as provided from the
// session table.
func (r PlayerRepo) DeleteSessionById(id string) error {
	_, err := r.pool.Exec(deleteSessionById, id)
	return err
}
