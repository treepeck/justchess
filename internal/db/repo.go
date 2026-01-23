package db

import (
	"database/sql"
)

// The following block of constants declares SQL queries,
// sorted by the tables they access.
const (
	// Player.

	insertPlayer = `INSERT INTO player (id, name, email, password_hash)
	VALUES (?, ?, ?, ?)`

	selectPlayerById = `SELECT * FROM player WHERE id = ?`

	selectPlayerByEmail = `SELECT * FROM player WHERE email = ?`

	selectPlayerBySessionId = `SELECT player.* FROM player INNER JOIN session
	ON player.id = session.player_id
	WHERE session.id = ? AND session.expires_at > NOW()`

	// Game.

	insertGame = `INSERT INTO game (id, white_id, black_id,
		time_control, time_bonus)
	VALUES (?, ?, ?, ?, ?)`

	// Select game by id excluding the abandoned games.
	selectGameById = `SELECT * FROM game WHERE id = ? AND termination != 1`

	// Session.

	insertSession = `INSERT INTO session (id, player_id) VALUES (?, ?)`

	selectSessionById = `SELECT * FROM session WHERE id = ? AND
	expires_at > NOW()`

	selectSessionsByPlayerId = `SELECT * FROM session WHERE player_id = ? AND
	expires_at > NOW()`

	deleteSessionById = `DELETE FROM session WHERE id = ?`
)

// Repo wraps the database connection pool and provides methods to make queries.
type Repo struct {
	pool *sql.DB
}

func NewRepo(pool *sql.DB) Repo { return Repo{pool: pool} }

// InsertPlayer inserts a single record into the player table, using the provided
// credentials.
func (r Repo) InsertPlayer(id, name, email string, passwordHash []byte) error {
	_, err := r.pool.Exec(insertPlayer, id, name, email, passwordHash)
	return err
}

// SelectPlayerById selects a single record with the same id as provided from the
// player table.
func (r Repo) SelectPlayerById(id string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerById, id)

	var p Player
	return p, row.Scan(&p.Id, &p.Name, &p.Rating, &p.RatingDeviation,
		&p.RatingVolatility, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
}

// SelectPlayerByEmail selects a single record with the same email as provided
// from the player table.
func (r Repo) SelectPlayerByEmail(email string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerByEmail, email)

	var p Player
	return p, row.Scan(&p.Id, &p.Name, &p.Rating, &p.RatingDeviation,
		&p.RatingVolatility, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
}

// SelectPlayerBySessionId selects a single player with the specified session_id.
// Expired sessions are omitted.
func (r Repo) SelectPlayerBySessionId(id string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerBySessionId, id)

	var p Player
	return p, row.Scan(&p.Id, &p.Name, &p.Rating, &p.RatingDeviation,
		&p.RatingVolatility, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
}

// InsertGame inserts a single record into the game table.
func (r Repo) InsertGame(id, whiteId, blackId string, control, bonus int) error {
	_, err := r.pool.Exec(insertGame, id, whiteId, blackId, control, bonus)
	return err
}

// SelectGameById selects a single record with the specified id from the game
// table.  Error is returned when the game doesn't exist.
func (r Repo) SelectGameById(id string) (Game, error) {
	row := r.pool.QueryRow(selectGameById, id)

	var g Game
	return g, row.Scan(&g.Id, &g.WhiteId, &g.BlackId, &g.Control, &g.Bonus,
		&g.Result, &g.Termination, &g.CreatedAt)
}

// InsertSession inserts a single record into the session table.
func (r Repo) InsertSession(id, playerId string) error {
	_, err := r.pool.Exec(insertSession, id, playerId)
	return err
}

// SelectSessionsByPlayerId selects a single record with the specified player_id
// from the session table.  Expired sessions are omitted.
func (r Repo) SelectSessionById(id string) (Session, error) {
	row := r.pool.QueryRow(selectSessionById, id)

	var s Session
	return s, row.Scan(&s.Id, &s.PlayerId, &s.CreatedAt, &s.ExpiresAt)
}

// SelectSessionsByPlayerId selects multiple records with the specified player_id
// from the session table.  Expired sessions are omitted.
func (r Repo) SelectSessionsByPlayerId(playerId string) ([]Session, error) {
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
func (r Repo) DeleteSessionById(id string) error {
	_, err := r.pool.Exec(deleteSessionById, id)
	return err
}
