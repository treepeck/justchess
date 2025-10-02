package db

import (
	"time"

	"database/sql"

	"github.com/treepeck/chego"
)

/*
The following block of constants declares the predifined queries to create and
modify database tables.
*/
const (
	// Player.

	insertPlayer = `INSERT INTO player (id, name, email, password_hash)
	VALUES (?, ?, ?, ?)`

	selectPlayerById = `SELECT * FROM player WHERE id = ?`

	selectPlayerByEmail = `SELECT * FROM player WHERE email = ?`

	// Session.

	insertSession = `INSERT INTO session (id, player_id) VALUES (?, ?)`

	selectPlayerBySessionId = `SELECT player.* FROM player INNER JOIN session
	ON player.id = session.player_id WHERE session.id = ?`

	deleteExpiredSessions = `DELETE FROM session WHERE expires_at < NOW()`
)

/*
Player represents a registered player.  Sensitive data, such as password hash and
email will not be encoded into a JSON.
*/
type Player struct {
	PasswordHash []byte    `json:"-"`
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

/*
Session stores the ID of the Authorizated player.  Each session has a 24-hour
lifecycle.  Expired sessions are deleted automatically when the player tries to
sign in.
*/
type Session struct {
	Id        string
	PlayerId  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

/*
Game represents the state of a single completed chess game.
*/
type Game struct {
	CompressedMoves []int32
	Id              string
	WhiteId         string
	BlackId         string
	TimeControl     int
	TimeBonus       int
	Result          chego.Result
	Winner          chego.Color
	CreatedAt       time.Time
}

/*
Repo wraps the database connection pool and provides methods to make queries.
*/
type Repo struct {
	pool *sql.DB
}

func NewRepo(pool *sql.DB) *Repo {
	return &Repo{pool: pool}
}

/*
InsertPlayer inserts a single record into the player table, using the provided
credentials.
*/
func (r *Repo) InsertPlayer(id, name, email string, passwordHash []byte) error {
	_, err := r.pool.Exec(insertPlayer, id, name, email, passwordHash)
	return err
}

/*
InsertSession inserts a single record into the session table.
*/
func (r *Repo) InsertSession(id, playerId string) error {
	_, err := r.pool.Exec(insertSession, id, playerId)
	return err
}

/*
SelectPlayerById selects a single record with the same id as provided from the player
table.
*/
func (r *Repo) SelectPlayerById(id string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerById, id)

	var p Player
	err := row.Scan(&p.Id, &p.Name, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
	return p, err
}

/*
SelectPlayerByEmail selects a single record with the same email as provided from the
player table.
*/
func (r *Repo) SelectPlayerByEmail(email string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerByEmail, email)

	var p Player
	err := row.Scan(&p.Id, &p.Name, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
	return p, err
}

/*
SelectPlayerBySessionId selects a single player record with the player_id,
similar to the one from session table.

NOTE: it may return player by exired session.  It's a caller's responsibility to
delete all expired sessions before calling this function.
*/
func (r *Repo) SelectPlayerBySessionId(id string) (Player, error) {
	row := r.pool.QueryRow(selectPlayerBySessionId, id)

	var p Player
	err := row.Scan(&p.Id, &p.Name, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
	return p, err
}

/*
DeleteExpiredSessions delets all expired records from the session table.
*/
func (r *Repo) DeleteExpiredSessions() error {
	_, err := r.pool.Exec(deleteExpiredSessions)
	return err
}
