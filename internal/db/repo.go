package db

import (
	"crypto/rand"
	"time"
)

type Player struct {
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"-"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func SelectPlayerByEmail(email string) (Player, error) {
	query := "SELECT * FROM player WHERE email = $1;"
	row := pool.QueryRow(query, email)

	var p Player
	err := row.Scan(&p.Id, &p.Name, &p.Email, &p.PasswordHash, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

// InsertPlayer returns an error if the record can't be inserted.
func InsertPlayer(name, email, pwdHash string) error {
	query := "INSERT INTO player (id, name, email, password_hash) VALUES ($1, $2, $3, $4);"

	_, err := pool.Exec(query, rand.Text(), name, email, pwdHash)
	return err
}

func DeletePlayerByName(name string) error {
	query := "DELETE FROM player WHERE name = $1;"

	_, err := pool.Exec(query, name)
	return err
}

func SelectSessionByPlayerId(pid string) (string, error) {
	query := "SELECT id FROM session WHERE player_id = $1;"
	row := pool.QueryRow(query, pid)

	var sid string
	err := row.Scan(&sid)
	return sid, err
}

// SelectPlayerIdBySessionId returns an error if the session is missing.
func SelectPlayerIdBySessionId(sid string) (string, error) {
	query := "SELECT player_id FROM session WHERE id = $1;"
	row := pool.QueryRow(query, sid)

	var pid string
	err := row.Scan(&pid)
	return pid, err
}

func DeleteExpiredSessions() error {
	query := "DELETE FROM session WHERE expires_at < now();"

	_, err := pool.Exec(query)
	return err
}

func InsertSession(sid string, pid string) error {
	query := "INSERT INTO session (id, player_id) VALUES ($1, $2);"

	_, err := pool.Exec(query, sid, pid)
	return err
}
