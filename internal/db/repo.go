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

	createPlayer = `CREATE TABLE IF NOT EXISTS player (
		id CHAR(12) PRIMARY KEY,
		name VARCHAR(60) NOT NULL UNIQUE,
		email VARCHAR(100) NOT NULL UNIQUE,
		password_hash VARCHAR(60) NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`

	insertPlayer = `INSERT INTO player (id, name, email, password_hash) VALUES
	(?, ?, ?, ?)`

	selectPlayerById = `SELECT * FROM player WHERE id = ?`

	selectPlayerByEmail = `SELECT * FROM player WHERE email = ?`

	// Session.

	createSession = `CREATE TABLE IF NOT EXISTS session (
		id CHAR(32) PRIMARY KEY,
		player_id CHAR(12) NOT NULL UNIQUE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME AS (created_at + INTERVAL 24 HOUR) STORED,
		FOREIGN KEY (player_id) REFERENCES player(id)
	)`

	insertSession = `INSERT INTO session (id, player_id) VALUES (?, ?)`

	selectSessionById = `SELECT * FROM session WHERE id = ?`

	deleteExpiredSessions = `DELETE FROM session WHERE expires_at < NOW()`

	// TODO: Game.
)

/*
Player represents a registered player.  Sensitive data, such as password hash and
email will not be encoded into a JSON.
*/
type Player struct {
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"-"`
	PasswordHash string    `json:"-"`
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
CreatePlayer creates the player table.
*/
func CreatePlayer(pool *sql.DB) error {
	_, err := pool.Exec(createPlayer)
	return err
}

/*
CreateSession creates the session table.
*/
func CreateSession(pool *sql.DB) error {
	_, err := pool.Exec(createSession)
	return err
}

/*
InsertPlayer inserts a single record into the player table, using the provided
credentials.
*/
func InsertPlayer(pool *sql.DB, id, name, email, passwordHash string) error {
	_, err := pool.Exec(insertPlayer, id, name, email, passwordHash)
	return err
}

/*
InsertSession inserts a single record into the session table.
*/
func InsertSession(pool *sql.DB, id, playerId string) error {
	_, err := pool.Exec(insertSession, id, playerId)
	return err
}

/*
SelectPlayerById selects a single record with the same id as provided from the player
table.
*/
func SelectPlayerById(pool *sql.DB, id string) (Player, error) {
	row := pool.QueryRow(selectPlayerById, id)

	var p Player
	err := row.Scan(&p.Id, &p.Name, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
	return p, err
}

/*
SelectPlayerByEmail selects a single record with the same email as provided from the
player table.
*/
func SelectPlayerByEmail(pool *sql.DB, email string) (Player, error) {
	row := pool.QueryRow(selectPlayerByEmail, email)

	var p Player
	err := row.Scan(&p.Id, &p.Name, &p.Email, &p.PasswordHash, &p.CreatedAt,
		&p.UpdatedAt)
	return p, err
}

/*
SelectSessionById selects a single record with the same id as provided from the
session table.

NOTE: It may return an expired session.  It's a caller's responsibility to
delete all expired sessions before calling this function.
*/
func SelectSessionById(pool *sql.DB, id string) (Session, error) {
	row := pool.QueryRow(selectSessionById, id)

	var s Session
	err := row.Scan(&s.Id, &s.PlayerId, &s.CreatedAt, &s.ExpiresAt)
	return s, err
}

/*
DeleteExpiredSessions delets all expired records from the session table.
*/
func DeleteExpiredSessions(pool *sql.DB) error {
	_, err := pool.Exec(deleteExpiredSessions)
	return err
}
